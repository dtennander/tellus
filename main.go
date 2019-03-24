// Main package of the tellus bot.
// Reads the configuration from disk and starts Tellus.
package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"tellus/tellus"
	"tellus/tellus/http"
)

const DefaultConfigurationLocation = "/config/tellus.yml"

// Deployment configuration that configures Tellus.
type Configuration struct {
	// The directory in which all repositories will be checked out and stored.
	RepositoryRootDirectory string  `yaml:"repositoryRootDirectory"`
	// The port on which the web server will be running.
	WebPort string                  `yaml:"webPort"`
	// Github configuration
	Github struct{
		// Information about the private key given by Github to authenticate Tellus.
		PrivateKey struct{
			// The location of the private key.
			Location string	        `yaml:"location"`
		}                           `yaml:"privateKey"`
		// The integration id given by Github.
		IntegrationId int           `yaml:"integrationId"`
		// The installation id given by Github.
		InstallationId int          `yaml:"installationId"`
	}  	                            `yaml:"github"`
}

func main() {
	log.Printf("Starting tellus!")
	config, err := getConfiguration()
	if err != nil {
		panic(err.Error())
	}
	log.Print(config)
	tellusClient, err := tellus.NewClient(
		config.Github.PrivateKey.Location,
		config.RepositoryRootDirectory,
		config.Github.IntegrationId,
		config.Github.InstallationId)
	if err != nil {
		log.Print(err.Error())
	}
	http.StartHttpServer(config.WebPort, tellusClient)
}

func getConfiguration() (*Configuration, error) {
	fileLocation := getConfigLocation()
	bytes, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return nil, err
	}
	log.Print(string(bytes))
	var config Configuration
	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getConfigLocation() string {
	fl := os.Getenv("CONFIG_FILE")
	if fl == "" {
		fl = DefaultConfigurationLocation
	}
	return fl
}
