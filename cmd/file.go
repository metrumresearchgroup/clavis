package cmd

import (
	"bytes"
	"os"
	"os/user"
	"text/template"

	"github.com/spf13/viper"
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
func (u RSConnectUser) TemplateSpec() (TemplateSpec, error) {

	//Create the Config object
	tspec := TemplateSpec{
		Organization: viper.GetString("organization"),
		Username:     viper.GetString("username"),
		PasswordFile: viper.GetString("location") + "/" + viper.GetString("file"),
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
func (t TemplateSpec) Write() error {

	//Write the Password File out
	password, err := os.Create(viper.GetString("location") + "/" + viper.GetString("file"))

	if err != nil {
		return err
	}
	defer password.Close()
	password.WriteString(GeneratedPassword + "\n")
	password.Chmod(0700)

	//Write the MOTD out
	content, err := t.Render()

	if err != nil {
		return err
	}

	motd, err := os.Create(viper.GetString("location") + "/.motd")

	if err != nil {
		return err
	}

	defer motd.Close()

	motd.Chmod(0700)
	motd.WriteString(content + "\n")

	//Updating Shell
	err = updateShellConfiguration()

	if err != nil {
		return err
	}

	return nil
}

func updateShellConfiguration() error {
	logos, err := user.Current()
	if err != nil {
		return err
	}

	config := logos.HomeDir + "/" + viper.GetString("shellconfig")

	f, err := os.OpenFile(config, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return err
	}

	defer f.Close()

	f.WriteString("cat " + logos.HomeDir + "/.motd")

	return nil
}
