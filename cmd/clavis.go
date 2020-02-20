package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	uuid "github.com/google/uuid"
	"github.com/mbndr/figlet4go"
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
}

// Clavis is the root level command
var Clavis = &cobra.Command{
	Use:   "clavis",
	Short: "Preparing and securing RSConnect",
	Long:  `This application serves to provision an initial RSConnect (Password backed) user. A password is generated and a templated file is inserted into the user's directory with those details for login purposes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {


		err := viper.Unmarshal(&ViperConfiguration)

		if ViperConfiguration.Debug{
			log.SetLevel(log.DebugLevel)
		}

		log.Debug("Successfully marshalled viper down")

		if err != nil {
			log.Fatalf("Failure unmarshalling viper contents")
		}

		//Check for config
		log.Debug("Looking for an existing configuration")
		existingConfig, err := readConfig()

		if err == nil && existingConfig.Completed {
			//Looks like there was actually a completed config file here.
			log.Info("A config already exists for this user. No work to be done")
			return
		}

		log.Debug("Attempting to create new user struct from Viper details")
		u := newRSConnectUser(ViperConfiguration)
		u, err = u.Create(ViperConfiguration)

		if err != nil {
			log.Errorf("An error occurred while creating the user in RSConnect: %s", err)
			return
		}

		//Config Storage
		newConfig := Configuration{
			Completed:              true,
			PasswordFile:           filepath.Join(ViperConfiguration.Location,ViperConfiguration.File),
			ShellConfigurationFile: ViperConfiguration.ShellConfig,
			Location:               ViperConfiguration.Location,
		}

		log.Debug("Writing the details of the clavis config")
		err = newConfig.Store()

		if err != nil {
			cmd.PrintErr(err)
			return
		}

		log.Info("Successfully provisioned user")
	},
}

func init() {

	user, err := user.Current()

	if err != nil {
		log.Fatalf("Unable to access current user details!. %s",err)
	}

	//Optionally specify a username besides the shell account name
	Clavis.PersistentFlags().StringP("username", "u", "", "The username to be utilized for both account creation and detail storage")
	Clavis.PersistentFlags().StringP("file", "f", ".rsconnectpassword", "The filename to use for writing the contents of the rsconnect password")
	//Optionally specify a location besides the current user home directory
	Clavis.PersistentFlags().StringP("location", "l", "", "The absolute path to a directory in which to store the key details")
	Clavis.PersistentFlags().StringP("email", "e", "user@domain.com", "The email to be used when generating the user")
	Clavis.PersistentFlags().StringP("name", "n", "", "The name of the user [First and last separated by space] we are provisioning")
	Clavis.PersistentFlags().StringP("organization", "o", "ThisCo", "The name of the organization used for setting up the template")
	Clavis.PersistentFlags().StringP("shell_config", "c", ".bashrc", "Defines the location of the shall profile / rc for manipulation")
	Clavis.PersistentFlags().BoolP("debug","d", false, "Whether or not to print debug information" )

	viper.SetDefault("user",user.Username)
	viper.SetDefault("location", user.HomeDir)

	viper.BindPFlag("username", Clavis.PersistentFlags().Lookup("username"))
	viper.BindPFlag("file", Clavis.PersistentFlags().Lookup("file"))
	viper.BindPFlag("location", Clavis.PersistentFlags().Lookup("location"))
	viper.BindPFlag("email", Clavis.PersistentFlags().Lookup("email"))
	viper.BindPFlag("name", Clavis.PersistentFlags().Lookup("name"))
	viper.BindPFlag("organization", Clavis.PersistentFlags().Lookup("organization"))
	viper.BindPFlag("shell_config", Clavis.PersistentFlags().Lookup("shell_config"))
	viper.BindPFlag("debug", Clavis.PersistentFlags().Lookup("debug"))
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
	if err != nil {
		return &http.Request{}, err
	}

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

	if resp.StatusCode > 500 || resp.StatusCode == http.StatusUnauthorized {
		return u, fmt.Errorf("We received an unexpected (%v) http response from the server", resp.StatusCode)
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
