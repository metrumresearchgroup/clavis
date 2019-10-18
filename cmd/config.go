package cmd

import (
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

//The naming of the file will always be .clavis.yml
const configFile string = ".clavis.yml"

//Configuration represents a way to store the results of a run in a hidden file in the home directory (.clavis) so that other commands can be run referencing it later.
type Configuration struct {
	Completed              bool   `yaml:"completed"`
	PasswordFile           string `yaml:"password_file"`
	ShellConfigurationFile string `yaml:"shellconfig"`
	Location               string `yaml:"location"`
}

//YAML will render the configuration to YAML structures
func (c Configuration) YAML() ([]byte, error) {
	return yaml.Marshal(c)
}

//Store is responsible for writing the configuration to the home directory for use later
func (c Configuration) Store() error {
	bytes, err := c.YAML()

	if err != nil {
		return err
	}

	user, err := user.Current()

	if err != nil {
		return err
	}

	file, err := os.Create(user.HomeDir + "/" + configFile)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(bytes)

	if err != nil {
		return err
	}

	return nil
}

func readConfig() (Configuration, error) {
	var conf Configuration
	user, err := user.Current()

	if err != nil {
		return Configuration{}, err
	}

	content, err := ioutil.ReadFile(user.HomeDir + "/" + configFile)

	if err != nil {
		return Configuration{}, err
	}

	err = yaml.Unmarshal(content, &conf)

	if err != nil {
		return Configuration{}, err
	}

	return conf, nil

}
