package cmd

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/user"
	"testing"
)

func Test_updateShellConfig(t *testing.T) {

	curuser, _ := user.Current()

	sampleContent := "oh hai meow"
	sampleLocation := "meow.txt"

	config := Configuration{
		ShellConfigurationFile: sampleLocation,
	}

	if !fileExists(sampleLocation) {
		ioutil.WriteFile(sampleLocation, []byte(sampleContent), 0644)
	}

	type args struct {
		user   *user.User
		config Configuration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Find and update file",
			args: args{
				user:   curuser,
				config: config,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateShellConfig(tt.args.user, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("updateShellConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHush(t *testing.T) {
	//First create motd file if it doesn't exist
	content := "meow"
	user, _ := user.Current()

	viper.Set("username", "dbreeden")


	if !fileExists(user.HomeDir + "/.motd") {
		ioutil.WriteFile(user.HomeDir+"/.motd", []byte(content), 0644)
	}

	cmd, _ := Hush.ExecuteC()
	err := hush(cmd)

	if err != nil {
		t.Errorf("Something went wrong: %s ", err.Error())
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
