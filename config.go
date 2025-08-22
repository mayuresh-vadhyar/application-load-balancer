package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type Config struct {
	Algorithm           string   `json:"algorithm"`
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	MaxUnhealthyChecks  int8     `json:"maxUnhealthyChecks"`
	Servers             []string `json:"servers"`
	Weights             []int    `json:"weights"`
}

var configOnce sync.Once
var config Config

func GetConfig() Config {
	configOnce.Do(func() {
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
	})

	return config
}
