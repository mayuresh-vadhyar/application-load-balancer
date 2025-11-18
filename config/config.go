package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type RateLimitConfig struct {
	Enable   bool   `json:"enable"`
	Strategy string `json:"strategy"`
	Limit    int    `json:"limit"`
	Window   string `json:"window"`
}

type HealthCheckConfig struct {
	MaxUnhealthyChecks  int8   `json:"maxUnhealthyChecks"`
	Interval 						string `json:"interval"`
	Cooldown 						string `json:"cooldown"`
	MaxRestart 					int8 `json:"maxRestart"`
}

type Config struct {
	Algorithm           string   `json:"algorithm"`
	Port                string   `json:"port"`
	DisableLogs         bool     `json:"disableLogs"`
	Servers             []string `json:"servers"`
	Weights             []int    `json:"weights"`
	RateLimit           RateLimitConfig
	HealthCheck         HealthCheckConfig
	RedisURL            string `json:"redis"`
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
