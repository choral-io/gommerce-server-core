package config

import (
	"os"

	yaml "github.com/goccy/go-yaml"
)

// LoadYamlConfig loads RootConfig from file.
// The path of the file is specified by environment variable GOMMERCE_CONFIG_PATH,
// If the environment variable is not set, it defaults to "./config/app-deploy.yaml".
func LoadYamlConfig() (RootConfig, error) {
	path, ok := os.LookupEnv("GOMMERCE_CONFIG_PATH")
	if !ok {
		path = "./config/app-deploy.yaml"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	raw = []byte(os.ExpandEnv(string(raw)))
	cfg := &rootConfig{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, err
	} else {
		cfg.rawData = raw
		return cfg, nil
	}
}
