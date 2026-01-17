package config

import (
	"encoding/json"
	"os"
	"time"
)

type ProxyConfig struct {
	Port            int           `json:"port"`
	Strategy        string        `json:"strategy"`
	HealthCheckFreq time.Duration `json:"health_check_frequency"`
}

func LoadConfig(filename string) (*ProxyConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	/*NOTE:
	* Regarding the health_check_frequency field, it will be written as a string in the config file
	* We can't just parse it into time.Duration as this can be risky,
	* The solution i found suitable is to get it as a string first, and then parse it as I wish
	 */
	//FIXME: If there is a better way I would like to know about it.
	var temp struct {
		Port            int    `json:"port"`
		Strategy        string `json:"strategy"`
		HealthCheckFreq string `json:"health_check_frequency"` // as you can see we are getting a string
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&temp)
	if err != nil {
		return nil, err
	}
	// Now we just need to convert the string
	duration, err := time.ParseDuration(temp.HealthCheckFreq)
	if err != nil {
		return nil, err
	}

	return &ProxyConfig{
		Port:            temp.Port,
		Strategy:        temp.Strategy,
		HealthCheckFreq: duration,
	}, nil

}
