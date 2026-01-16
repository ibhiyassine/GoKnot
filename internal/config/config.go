package config

import "time"

type ProxyConfig struct {
	Port            int    `json:"port"`
	Strategy        string `json:"strategy"`
	HealthCheckFreq time.Duration
}
