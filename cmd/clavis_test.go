package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func Test_newRSConnectUser(t *testing.T) {

	type sourcing struct {
		emails    []string
		names     []string
		usernames []string
	}

	s := sourcing{
		emails: []string{
			"john@doe.com",
			"jane@doe.com",
			"jane@doe.com",
		},
		names: []string{
			"John Doe",
			"Jane",
			"Jane",
		},
		usernames: []string{
			"johnd",
			"",
			"",
		},
	}

	tests := []struct {
		name string
		want RSConnectUser
	}{
		{
			name: "Full Name Parsed",
			want: RSConnectUser{
				Email:          "john@doe.com",
				FirstName:      "John",
				LastName:       "Doe",
				Password:       "123456",
				SetOwnPassword: false,
				Username:       "johnd",
			},
		},
		{
			name: "Atypical full name",
			want: RSConnectUser{
				Email:          "jane@doe.com",
				FirstName:      "Jane",
				Password:       "123456",
				SetOwnPassword: false,
			},
		},
		{
			name: "No provided Username",
			want: RSConnectUser{
				Email:          "jane@doe.com",
				FirstName:      "Jane",
				Password:       "123456",
				SetOwnPassword: false,
			},
		},
	}
	for key, tt := range tests {

		//Lets Prep Viper
		viper.Set("username", s.usernames[key])
		viper.Set("email", s.emails[key])
		viper.Set("name", s.names[key])

		var vc ViperConfig
		viper.Unmarshal(&vc)

		t.Run(tt.name, func(t *testing.T) {
			got := newRSConnectUser(vc)
			got.Password = "123456"

			got.Username = s.usernames[key]

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run %v newRSConnectUser() = %v, want %v", key, got, tt.want)
			}
		})
	}
}

func TestRSConnectUser_JSON(t *testing.T) {

	initial := RSConnectUser{
		Email:          "john@doe.com",
		FirstName:      "John",
		LastName:       "Doe",
		Password:       "123456",
		SetOwnPassword: false,
		Username:       "johnd",
	}

	initialSerialized, _ := json.Marshal(initial)
	println(string(initialSerialized))

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
		want    []byte
		wantErr bool
	}{
		{
			name: "JSON Generation Successful",
			fields: fields{
				Email:          "john@doe.com",
				FirstName:      "John",
				LastName:       "Doe",
				Password:       "123456",
				SetOwnPassword: false,
				Username:       "johnd",
			},
			want:    []byte(`{"email":"john@doe.com","first_name":"John","last_name":"Doe","password":"123456","user_must_set_password":false,"username":"johnd"}`),
			wantErr: false,
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
			got, err := u.JSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("RSConnectUser.JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RSConnectUser.JSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRSConnectUser_Request(t *testing.T) {

	ruser := RSConnectUser{
		Email:          "john@doe.com",
		FirstName:      "John",
		LastName:       "Doe",
		Password:       "123456",
		SetOwnPassword: false,
	}

	ruserSerialized, _ := json.Marshal(ruser)

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
		wantErr bool
	}{
		{
			name: "General req creation",
			fields: fields{
				Email:          ruser.Email,
				FirstName:      ruser.FirstName,
				LastName:       ruser.LastName,
				Password:       ruser.Password,
				SetOwnPassword: ruser.SetOwnPassword,
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
			got, err := u.Request()
			if (err != nil) != tt.wantErr {
				t.Errorf("RSConnectUser.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			//Does the body match the serialized content?
			bodyContent, _ := ioutil.ReadAll(got.Body)

			if !reflect.DeepEqual(bodyContent, ruserSerialized) {
				t.Errorf("The content was not encoded or set as expected")
			}

			//Is the request marked as json content?
			if got.Header.Get("Content-Type") != "application/json" {
				t.Errorf("There was a failure in setting the encoding on this request")
			}
		})
	}
}

func TestRSConnectUser_Create(t *testing.T) {
	//Gonna be more freeform with this bad boy.

	viper.Set("location", "/tmp")
	viper.Set("file", "testfile")

	var vc ViperConfig
	viper.Unmarshal(&vc)

	vc.Prepare()

	var responsecode int = 401

	u := RSConnectUser{
		Email:          "john@doe.com",
		FirstName:      "John",
		LastName:       "Doe",
		Password:       "123456",
		SetOwnPassword: false,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		new := u
		new.ActiveTime = "now"
		new.CreatedTime = "now"
		new.UpdatedTime = "now"
		new.GUID = "abc-def-ghi"
		new.UserRole = "administrator"

		serialized, _ := new.JSON()
		w.Header().Add("content-type", "application/json")
		w.Write(serialized)
	}))

	defer srv.Close()

	//Remap value so that we can hit the test server.
	userAPIURL = srv.URL

	newU, err := u.Create(vc)

	if err != nil {
		t.Errorf("Failed for some reason %s", err.Error())
	}

	if newU.CreatedTime == "" {
		t.Errorf("Failed to get a creation stamp back from the server")
	}

	badsrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responsecode)
		w.Header().Add("content-type", "application/json")
		w.Write([]byte(""))
	}))

	defer badsrv.Close()

	userAPIURL = badsrv.URL

	_, err = u.Create(vc)

	if err == nil {
		t.Errorf("We should have gotten an error back due to a 401")
	}

	responsecode = 501

	_, err = u.Create(vc)

	if err == nil {
		t.Errorf("We should have received an error back to a 501")
	}

	//Now we'll test a failure mode of a 401 /

}

func TestCommandExecution(t *testing.T) {
	//Gonna be more freeform with this bad boy.

	viper.Set("location", "/tmp")
	viper.Set("file", "testfile")

	viper.Set("organization", "thisco")
	viper.Set("name", "this guy")
	viper.Set("email", "this@guy.com")

	u := RSConnectUser{
		Email:          "john@doe.com",
		FirstName:      "John",
		LastName:       "Doe",
		Password:       "123456",
		SetOwnPassword: false,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		new := u
		new.ActiveTime = "now"
		new.CreatedTime = "now"
		new.UpdatedTime = "now"
		new.GUID = "abc-def-ghi"
		new.UserRole = "administrator"

		serialized, _ := new.JSON()
		w.Header().Add("content-type", "application/json")
		w.Write(serialized)
	}))

	defer srv.Close()

	//Remap value so that we can hit the test server.
	userAPIURL = srv.URL

	Clavis.Execute()

}
