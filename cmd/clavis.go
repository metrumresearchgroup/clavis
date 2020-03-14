package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/afero"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	uuid "github.com/google/uuid"
	"github.com/mbndr/figlet4go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//GeneratedPassword is the GUID generated for the purpose of one-time authentication
var GeneratedPassword string

var userAPIURL string = "http://localhost:3939/__api__/v1/users"

//RSConnectUser is the structure defining an RSConnect User
type RSConnectUser struct {
	//Components required for transmission
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Password       string `json:"password"`
	SetOwnPassword bool   `json:"user_must_set_password"`
	Username       string `json:"username"`
	//Additional components for serializing response
	ActiveTime  string `json:"active_time,omitempty"`
	Confirmed   bool   `json:"confirmed,omitempty"`
	CreatedTime string `json:"created_time,omitempty"`
	GUID        string `json:"guid,omitempty"`
	Locked      bool   `json:"locked,omitempty"`
	UpdatedTime string `json:"updated_time,omitempty"`
	UserRole    string `json:"user_role,omitempty"`
}

var (
	ViperConfiguration ViperConfig
)

type ViperConfig struct {
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	File string `mapstructure:"file" json:"file" yaml:"file"`
	Location string `mapstructure:"location" json:"location" yaml:"location"`
	Email string `mapstructure:"email" json:"email" yaml:"email"`
	Name string `mapstructure:"name" json:"name" yaml:"name"`
	Organization string `mapstructure:"organization" json:"organization" yaml:"organization"`
	ShellConfig string `mapstructure:"shell_config" json:"shell_config" yaml:"shell_config"`
	Debug bool `mapstructure:"debug" json:"debug" yaml:"debug"`
	CreateMOTD bool `mapstructure:"create_motd" json:"create_motd" yaml:"create_motd"`
	UID int `mapstructure:"-" json:"-" yaml:"-"`
	GID int `mapstructure:"-" json:"-" yaml:"-"`
	HomeDir string `mapstructure:"-" json:"-" yaml:"-"`
}

// Clavis is the root level command
var Clavis = &cobra.Command{
	Use:   "clavis",
	Short: "Preparing and securing RSConnect",
	Long:  `This utility provisions an admin user in RSConnect. 
A password is generated and effort is made to surface those details to the user for login purposes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		err := viper.Unmarshal(&ViperConfiguration)
		if err != nil {
			log.Fatalf("Failure unmarshalling viper contents")
		}

		log.Debug("Successfully marshalled viper down")
		if ViperConfiguration.Debug{
			log.SetLevel(log.DebugLevel)
		}

		log.Debug("Setting username")
		uname := viper.GetString("username")
		log.Debug("Looking up user details")
		userdetails, err := user.Lookup(uname)
		if err != nil {
			log.Fatalf("Could not get current user details for some reason: %s", err)
		}
		if viper.GetString("location") == "" {
			ViperConfiguration.Location = userdetails.HomeDir
		}
		ViperConfiguration.Username = userdetails.Username
		ViperConfiguration.UID, _ = strconv.Atoi(userdetails.Uid)
		ViperConfiguration.GID, _ = strconv.Atoi(userdetails.Gid)
		ViperConfiguration.HomeDir = userdetails.HomeDir

		log.Debugf("Located username as %s", ViperConfiguration.Username)

		//Check for config
		log.Debug("Checking for an existing configuration")
		existingConfig, err := readConfig(ViperConfiguration.Username)
		if err == nil && existingConfig.Completed {
			//Looks like there was actually a completed config file here.
			log.Info("A config already exists for this user. No work to be done")
		}

		//Lastly, let's check to see if the rsconnectpassword file exists in the user's home directory
		rspassLocation := filepath.Join(userdetails.HomeDir,ViperConfiguration.File)

		if ok, _ := afero.Exists(afero.NewOsFs(),rspassLocation); ok {
			log.Fatalf("An existing RSConnect password file was located at %s. Exiting", rspassLocation)
		}

		log.Debug("Attempting to create new user struct from Viper details")
		u := newRSConnectUser(ViperConfiguration)
		u, err = u.Create(ViperConfiguration)
		if err != nil {
			log.Fatalf("An error occurred while creating the user in RSConnect: %s", err)
		}

		//Config Storage
		newConfig := Configuration{
			Completed:              true,
			PasswordFile:           filepath.Join(ViperConfiguration.Location,ViperConfiguration.File),
			ShellConfigurationFile: ViperConfiguration.ShellConfig,
			Location:               ViperConfiguration.Location,
		}

		log.Debug("Writing the details of the clavis config")
		err = newConfig.Store(ViperConfiguration.Username)
		if err != nil {
			log.Errorf("error writing config to disk: %v\n", err)
		}

		log.Info("Successfully provisioned user")
	},
}

func init() {
	//Optionally specify a username besides the shell account name
	Clavis.PersistentFlags().StringP("username", "u", "",
		"The username to be utilized for both account creation and detail storage")
	Clavis.PersistentFlags().StringP("file", "f", ".rsconnectpassword", "The filename to use for writing the contents of the rsconnect password")
	//Optionally specify a location besides the current user home directory
	Clavis.PersistentFlags().StringP("location", "l", "", "The absolute path to a directory in which to store the key details")
	Clavis.PersistentFlags().StringP("email", "e", "", "The email to be used when generating the user")
	Clavis.PersistentFlags().StringP("name", "n", "", "The name of the user [First and last separated by space] we are provisioning")
	Clavis.PersistentFlags().StringP("organization", "o", "", "The name of the organization used for setting up the template")
	Clavis.PersistentFlags().StringP("shell_config", "c", ".bashrc",
		"Defines the location of the shell profile / rc for manipulation")
	Clavis.PersistentFlags().BoolP("debug","d", false, "Whether or not to print debug information" )
	Clavis.PersistentFlags().BoolP("create_motd", "s", true, "Whether or not to create the motd file for the user")
	Clavis.MarkFlagRequired("email")
	Clavis.MarkFlagRequired("name")
	Clavis.MarkFlagRequired("organization")


	viper.BindPFlag("username", Clavis.PersistentFlags().Lookup("username"))
	viper.BindPFlag("file", Clavis.PersistentFlags().Lookup("file"))
	viper.BindPFlag("location", Clavis.PersistentFlags().Lookup("location"))
	viper.BindPFlag("email", Clavis.PersistentFlags().Lookup("email"))
	viper.BindPFlag("name", Clavis.PersistentFlags().Lookup("name"))
	viper.BindPFlag("organization", Clavis.PersistentFlags().Lookup("organization"))
	viper.BindPFlag("shell_config", Clavis.PersistentFlags().Lookup("shell_config"))
	viper.BindPFlag("debug", Clavis.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("create_motd", Clavis.PersistentFlags().Lookup("create_motd"))
}

func newRSConnectUser(config ViperConfig) RSConnectUser {
	GeneratedPassword = uuid.New().String()

	namePieces := strings.Fields(config.Name)

	ucr := RSConnectUser{
		Username:       config.Username,
		Email:          config.Email,
		Password:       GeneratedPassword,
		FirstName:      config.Username,
		LastName:       config.Username,
		SetOwnPassword: false,
	}

	//Processing name details and handling of nulls to default
	if len(namePieces) > 0 {
		//This would be a first \s last name
		if len(namePieces) == 2 {
			ucr.FirstName = namePieces[0]
			ucr.LastName = namePieces[1]
		}

		//Leave the last name defaulted to the username
		if len(namePieces) == 1 {
			ucr.FirstName = namePieces[0]
		}
	}

	return ucr
}

//JSON returns the json representative string for the object
func (u RSConnectUser) JSON() ([]byte, error) {
	return json.Marshal(u)
}

//Request generates the HTTP request to be made
func (u RSConnectUser) Request() (*http.Request, error) {
	body, err := u.JSON()

	log.Debugf("Body we're sending is %s", body)
	if err != nil {
		return &http.Request{}, err
	}

	log.Debugf("Currently sending details to %s", userAPIURL)
	req, err := http.NewRequest("POST", userAPIURL, bytes.NewBuffer(body))

	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

//Create performs the request to the RS Connect server
func (u RSConnectUser) Create(config ViperConfig) (RSConnectUser, error) {
	req, err := u.Request()
	if err != nil {
		return u, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return u, err
	}

	log.Debugf("Status code we received back is %d", resp.StatusCode)

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		return u, fmt.Errorf("we received an unexpected (%d) http response from the server. Detais from the response are %s", resp.StatusCode, content)
	}

	//Just to make sure we close the thing.
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return u, err
	}

	var respUser RSConnectUser
	err = json.Unmarshal(content, &respUser)

	if err != nil {
		return u, err
	}

	//Trigger the File creation and update of bashrc
	ts, err := u.TemplateSpec(config)

	if err != nil {
		return u, err
	}

	err = ts.Write(config)

	if err != nil {
		return u, err
	}

	return respUser, nil
}

//GetFiglyWithIt Produces figlet output
func GetFiglyWithIt(input string) (string, error) {
	ascii := figlet4go.NewAsciiRender()

	// Adding the colors to RenderOptions
	options := figlet4go.NewRenderOptions()
	options.FontColor = []figlet4go.Color{
		figlet4go.ColorGreen,
	}

	return ascii.RenderOpts(input, options)
}
