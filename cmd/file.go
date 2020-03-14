package cmd

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

//TemplateSpec is a struct used for the outputTemplate file template
type TemplateSpec struct {
	OrganizationFiglet string
	Organization       string
	Username           string
	PasswordFile       string
}

const outputTemplate string = `
{{ if .OrganizationFiglet }}
{{- .OrganizationFiglet -}}
{{ end }}

Welcome to {{ .Organization }}. RSConnect has been provisioned on this system and your user, {{ .Username }}, has been provisioned as an administrator successfully! 
Your RSConnect password has been written to {{ .PasswordFile }} , but you should change it as quickly as soon as you login. 

If you'd like to stop seeing this message, just issue the following command:

clavis hush

This will remove the MOTD file displaying this message as well as update your bash configuration to not attempt to display it.


Enjoy {{ .Organization }}, and enjoy RSConnect!
`

//TemplateSpec creates a TemplateSpec from the RSConnect User
func (u RSConnectUser) TemplateSpec(config ViperConfig) (TemplateSpec, error) {

	//Create the Config object
	tspec := TemplateSpec{
		Organization: config.Organization,
		Username:     config.Username,
		PasswordFile: filepath.Join(config.Location , config.File),
	}

	fig, err := GetFiglyWithIt(tspec.Organization)

	if err != nil {
		return TemplateSpec{}, err
	}

	tspec.OrganizationFiglet = fig

	return tspec, nil
}

//Render is used to generate content to the template structure
func (t TemplateSpec) Render() (string, error) {
	template := template.New("file")
	template.Parse(outputTemplate)
	var rendered bytes.Buffer
	err := template.Execute(&rendered, t)

	if err != nil {
		return "", err
	}

	return rendered.String(), nil
}

//Write will write the rendered content down to the desired File
func (t TemplateSpec) Write(config ViperConfig) error {
	log.WithFields(log.Fields{
		"user" : config.UserDetails.Username,
		"location" : config.Location,
	}).Debug("Getting ready to write template spec to directory")
	//Write the Password File out

	password, err := os.Create(filepath.Join(config.Location,config.File))

	if err != nil {
		return err
	}
	defer password.Close()
	password.WriteString(GeneratedPassword + "\n")
	password.Chmod(0700)

	userid, err := strconv.Atoi(config.UserDetails.Uid)

	if err != nil {
		return err
	}

	guid, err := strconv.Atoi(config.UserDetails.Gid)

	if err != nil {
		return err
	}

	//Make sure the file is owned by the user we have looked up
	err = password.Chown(userid,guid)

	if err != nil {
		return err
	}

	log.Debug("Password file is taken care of")

	//Write the MOTD out
	content, err := t.Render()

	if err != nil {
		return err
	}

	if config.CreateMOTD {
		log.Debugf("Based on configuration, creating MOTD file in %s", config.UserDetails.HomeDir)
		motd, err := os.Create(filepath.Join(config.UserDetails.HomeDir,".motd"))

		if err != nil {
			return err
		}

		defer motd.Close()

		motd.Chmod(0700)
		motd.WriteString(content + "\n")

		err = motd.Chown(userid,guid)

		if err != nil {
			return err
		}

		//Updating Shell
		log.Debug("Updating shell / shell config")
		err = updateShellConfiguration(config)

		if err != nil {
			return err
		}
	}
	return nil
}

func updateShellConfiguration(vconfig ViperConfig) error {

	config := filepath.Join(vconfig.UserDetails.HomeDir,vconfig.ShellConfig)

	f, err := os.OpenFile(config, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer f.Close()

	f.WriteString("cat " + vconfig.UserDetails.HomeDir + "/.motd")

	return nil
}
