package cmd

import (
	"bytes"
	"reflect"
	"testing"
	"text/template"

	"github.com/spf13/viper"
)

func TestRSConnectUser_TemplateSpec(t *testing.T) {

	viper.Set("username", "joed")
	viper.Set("location", "/home/joed")
	viper.Set("file", ".rsconnectpassword")
	viper.Set("organization","ThisCo")

	var vc ViperConfig
	viper.Unmarshal(&vc)

	fig, _ := GetFiglyWithIt("ThisCo")

	type fields struct {
		Email          string
		FirstName      string
		LastName       string
		Password       string
		SetOwnPassword bool
		Username       string
		ActiveTime     string
		Confirmed      bool
		CreatedTime    string
		GUID           string
		Locked         bool
		UpdatedTime    string
		UserRole       string
	}
	tests := []struct {
		name    string
		fields  fields
		want    TemplateSpec
		wantErr bool
	}{
		{
			name: "Should setup fine",
			fields: fields{
				Email:          "email@gmail.com",
				FirstName:      "John",
				LastName:       "Doe",
				Password:       "123456",
				SetOwnPassword: false,
			},
			want: TemplateSpec{
				OrganizationFiglet: fig,
				Username:           viper.GetString("username"),
				PasswordFile:       viper.GetString("location") + "/" + viper.GetString("file"),
				Organization:       "ThisCo",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := RSConnectUser{
				Email:          tt.fields.Email,
				FirstName:      tt.fields.FirstName,
				LastName:       tt.fields.LastName,
				Password:       tt.fields.Password,
				SetOwnPassword: tt.fields.SetOwnPassword,
				Username:       tt.fields.Username,
				ActiveTime:     tt.fields.ActiveTime,
				Confirmed:      tt.fields.Confirmed,
				CreatedTime:    tt.fields.CreatedTime,
				GUID:           tt.fields.GUID,
				Locked:         tt.fields.Locked,
				UpdatedTime:    tt.fields.UpdatedTime,
				UserRole:       tt.fields.UserRole,
			}
			got, err := u.TemplateSpec(vc)
			if (err != nil) != tt.wantErr {
				t.Errorf("RSConnectUser.TemplateSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RSConnectUser.TemplateSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateSpec_Render(t *testing.T) {

	template := template.New("test")
	template.Parse(outputTemplate)

	tspec := TemplateSpec{
		OrganizationFiglet: "Metworx",
		Organization:       "Metworx",
		Username:           "darrellb",
		PasswordFile:       "/data/home/darrellb/.rsconnectpassword",
	}

	var buf bytes.Buffer
	template.Execute(&buf, tspec)

	type fields struct {
		OrganizationFiglet string
		Organization       string
		Username           string
		PasswordFile       string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Known rendering",
			fields: fields{
				OrganizationFiglet: "Metworx",
				Organization:       "Metworx",
				Username:           "darrellb",
				PasswordFile:       "/data/home/darrellb/.rsconnectpassword",
			},
			want:    buf.String(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := TemplateSpec{
				OrganizationFiglet: tt.fields.OrganizationFiglet,
				Organization:       tt.fields.Organization,
				Username:           tt.fields.Username,
				PasswordFile:       tt.fields.PasswordFile,
			}
			got, err := r.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("TemplateSpec.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TemplateSpec.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateSpec_Write(t *testing.T) {

	locations := []string{
		"/tmp",
		"/root",
	}

	filenames := []string{
		".motd",
		".motd",
	}

	type fields struct {
		OrganizationFiglet string
		Organization       string
		Username           string
		PasswordFile       string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Accessible file",
			fields: fields{
				OrganizationFiglet: "Metworx",
				Organization:       "Metworx",
				Username:           "darrellb",
				PasswordFile:       ".rsconnectpassword",
			},
			wantErr: false,
		},
		{
			name: "Unable to access file",
			fields: fields{
				OrganizationFiglet: "Metworx",
				Organization:       "Metworx",
				Username:           "darrellb",
				PasswordFile:       ".rsconnectpassword",
			},
			wantErr: true,
		},
	}
	for key, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			viper.Set("location", locations[key])
			viper.Set("filename", filenames[key])

			var vc ViperConfig
			vc.Prepare()
			viper.Unmarshal(&vc)

			r := TemplateSpec{
				OrganizationFiglet: tt.fields.OrganizationFiglet,
				Organization:       tt.fields.Organization,
				Username:           tt.fields.Username,
				PasswordFile:       tt.fields.PasswordFile,
			}
			if err := r.Write(vc); (err != nil) != tt.wantErr {
				t.Errorf("TemplateSpec.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
