package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
	Weights             []int    `json:"weights"`
}

func GetConfig() Config {
	var config Config

	// Read file
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal JSON into config struct
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	return config
}
