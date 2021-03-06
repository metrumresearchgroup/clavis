package cmd

import (
	"os/user"
	"testing"
)

func TestConfiguration_Store(t *testing.T) {

	user, _ := user.Current()

	vc := ViperConfig{
		Username:     user.Username,
		File:         "",
		Location:     user.HomeDir,
		Email:        "dukeofubuntu@gmail.com",
		Name:         "Darrell Breeden",
		Organization: "Metrum Research Group",
		ShellConfig:  "",
		Debug:        true,
		CreateMOTD:   true,
		UserDetails:  user,
	}

	type fields struct {
		Completed              bool
		PasswordFile           string
		ShellConfigurationFile string
		Location               string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Storage of yaml config",
			fields: fields{
				Completed:              true,
				PasswordFile:           user.HomeDir + "/.rsconnectpassword",
				ShellConfigurationFile: user.HomeDir + "/.bashrc",
				Location:               user.HomeDir,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Configuration{
				Completed:              tt.fields.Completed,
				PasswordFile:           tt.fields.PasswordFile,
				ShellConfigurationFile: tt.fields.ShellConfigurationFile,
				Location:               tt.fields.Location,
			}
			if err := c.Store(vc.UserDetails); (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
