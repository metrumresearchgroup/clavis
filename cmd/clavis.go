package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/user"
	"strings"

	uuid "github.com/google/uuid"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//GeneratedPassword is the GUID generated for the purpose of one-time authentication
var GeneratedPassword string

const userAPIURL string = "http://localhost:3939/__api__/v1/users"

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

// Clavis is the root level command
var Clavis = &cobra.Command{
	Use:   "Provisioning and user preparation tool",
	Short: "Preparing and securing",
	Long:  `This application serves to provision an initial RSConnect (Password backed) user. A password is generated and a templated file is inserted into the user's directory with those details for login purposes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		logos, err := user.Current()

		if err != nil {
			cmd.PrintErr("Unable to get current user details: ", err)
			return
		}

		//Handle empty /default settings
		if viper.GetString("username") == "" {
			viper.Set("username", logos.Username)
		}

		if viper.GetString("location") == "" {
			viper.Set("location", logos.HomeDir)
		}

		u := newRSConnectUser()
		u, err = u.Create()

		if err != nil {
			cmd.PrintErr(err)
			return
		}

		cmd.Println("Successfully provisioned user")
	},
}

func init() {
	//Optionally specify a username besides the shell account name
	Clavis.PersistentFlags().StringP("username", "u", "", "The username to be utilized for both account creation and detail storage")
	Clavis.PersistentFlags().StringP("file", "f", ".rsconnectpassword", "The filename to use for writing the contents of the rsconnect password")
	//Optionally specify a location besides the current user home directory
	Clavis.PersistentFlags().StringP("location", "l", "", "The absolute path to a directory in which to store the key details")
	Clavis.PersistentFlags().StringP("email", "e", "user@domain.com", "The email to be used when generating the user")
	Clavis.PersistentFlags().StringP("name", "n", "", "The name of the user [First and last separated by space] we are provisioning")
	Clavis.PersistentFlags().StringP("organization", "o", "ThisCo", "The name of the organization used for setting up the template")
	Clavis.PersistentFlags().StringP("shellconfig", "c", ".bashrc", "Defines the location of the shall profile / rc for manipulation")

	viper.BindPFlag("username", Clavis.PersistentFlags().Lookup("username"))
	viper.BindPFlag("file", Clavis.PersistentFlags().Lookup("file"))
	viper.BindPFlag("location", Clavis.PersistentFlags().Lookup("location"))
	viper.BindPFlag("email", Clavis.PersistentFlags().Lookup("email"))
	viper.BindPFlag("name", Clavis.PersistentFlags().Lookup("name"))
	viper.BindPFlag("organization", Clavis.PersistentFlags().Lookup("organization"))
	viper.BindPFlag("shellconfig", Clavis.PersistentFlags().Lookup("shellconfig"))
}

func newRSConnectUser() RSConnectUser {
	GeneratedPassword = uuid.New().String()

	namePieces := strings.Fields(viper.GetString("name"))

	ucr := RSConnectUser{
		Username:       viper.GetString("username"),
		Email:          viper.GetString("email"),
		Password:       GeneratedPassword,
		FirstName:      viper.GetString("username"),
		LastName:       viper.GetString("username"),
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
func (u RSConnectUser) Create() (RSConnectUser, error) {
	req, err := u.Request()
	if err != nil {
		return u, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return u, err
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
	ts, err := u.TemplateSpec()

	if err != nil {
		return u, err
	}

	err = ts.Write()

	if err != nil {
		return u, err
	}

	return u, nil
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
