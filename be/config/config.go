package config

import (
	"encoding/json"
	"os"
)

type config struct {
	ApiEthKey   string `json:"api_eth_key,omitempty"`
	S3Endpoint  string `json:"s3_endpoint,omitempty"`
	S3AccessKey string `json:"s3_access_key,omitempty"`
	S3SecretKey string `json:"s3_secret_key,omitempty"`
}

var defaultConfig = &config{}

func GetConfig() *config {
	return defaultConfig
}

func LoadConfig(pathCfg string) error {
	data, err := os.ReadFile(pathCfg)
	if err != nil {
		return err
	}
	json.Unmarshal(data, defaultConfig)
	return nil
}
