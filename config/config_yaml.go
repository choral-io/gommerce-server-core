package config

import (
	"bytes"
	"fmt"
	"os"

	yaml "github.com/goccy/go-yaml"
)

type yamlRootConfig struct {
	rootConfig

	yamlDoc []byte
}

func (c *yamlRootConfig) GetValue(path string, value any) error {
	if len(c.yamlDoc) == 0 {
		return ErrConfigNotLoaded
	}

	if (len(path) == 0) || (value == nil) {
		return fmt.Errorf("invalid arguments: path=%q, value=%T", path, value)
	}

	pyp, err := yaml.PathString(path) // parsed YAML path
	if err != nil {
		return fmt.Errorf("failed to parse YAML path %q: %w", path, err)
	}

	err = pyp.Read(bytes.NewReader(c.yamlDoc), value)
	if err != nil {
		return fmt.Errorf("failed to read value at path %q: %w", path, err)
	}

	return nil
}

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
	cfg := &yamlRootConfig{}
	if err := yaml.Unmarshal(raw, &cfg.rootConfig); err != nil {
		return nil, err
	} else {
		cfg.yamlDoc = raw
		return cfg, nil
	}
}
