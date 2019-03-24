package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"tellus/tellus"
	"tellus/tellus/http"
)

var PORT = os.Getenv("PORT")

const DefaultConfigurationLocation = "/config/tellus.yml"

type Configuration struct {
	RepositoryRootDirectory string 	`yaml:"repositoryRootDirectory"`
	Github struct{
		PrivateKey struct{
			Location string			`yaml:"location"`
		} 						    `yaml:"privateKey"`
		IntegrationId int			`yaml:"integrationId"`
		InstallationId int			`yaml:"installationId"`
	}  `yaml:"github"`
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
	http.ServeHttpClient(PORT, tellusClient)
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
