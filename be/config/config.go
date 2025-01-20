package config

import (
	"encoding/json"
	"os"

	"go.uber.org/zap"
)

type GoogleConsole struct {
	ID          string `yaml:"id" json:"id,omitempty"`
	Secret      string `yaml:"secret" json:"secret,omitempty"`
	CallbackSSO string `yaml:"callback_sso" json:"callback_sso,omitempty"`
	RedirectURI string `yaml:"redirect_uri" json:"redirect_uri,omitempty"`
}

type config struct {
	ApiEthKey             string        `json:"api_eth_key,omitempty"`
	S3Endpoint            string        `json:"s3_endpoint,omitempty"`
	S3AccessKey           string        `json:"s3_access_key,omitempty"`
	S3SecretKey           string        `json:"s3_secret_key,omitempty"`
	PostgressDsn          string        `json:"postgress_dsn,omitempty"`
	GoogleConsole         GoogleConsole `json:"google_console,omitempty"`
	GoogleConsoleCustomer GoogleConsole `json:"google_console_customer,omitempty"`
	TrustOrigin           []string      `json:"trust_origin,omitempty"`
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
	if len(defaultConfig.GoogleConsole.CallbackSSO) == 0 {
		zap.L().Warn("GoogleConsole CallbackSSO is empty")
	}
	if len(defaultConfig.GoogleConsole.RedirectURI) == 0 {
		zap.L().Warn("GoogleConsole RedirectURI is empty")
	}
	return nil
}
