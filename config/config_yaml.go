package config

import (
	"os"

	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

// LoadYamlConfig loads RootConfig from file.
// The path of the file is specified by environment variable GOMMERCE_CONFIG_PATH,
// If the environment variable is not set, it defaults to "./config/app-deploy.yaml".
func LoadYamlConfig() (RootConfig, error) {
	path, ok := os.LookupEnv("GOMMERCE_CONFIG_PATH")
	if !ok {
		path = "./config/app-deploy.yaml"
	}
	txt, err := os.ReadFile(path)
	txt = []byte(os.ExpandEnv(string(txt)))
	if err != nil {
		return nil, err
	}
	cfg := &rootConfig{}
	doc := new(any)
	if err := yaml.Unmarshal(txt, cfg); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(txt, doc); err != nil {
		return nil, err
	} else {
		cfg.yamlDoc = *doc
		return cfg, nil
	}
}
