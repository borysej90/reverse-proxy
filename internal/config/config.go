package config

import (
	"gopkg.in/yaml.v3"
	"io"
)

type path struct {
	Location        string `yaml:"location"`
	Target          string `yaml:"target"`
	ConnectionLimit int    `yaml:"connection_limit"`
	DropOverLimit   bool   `yaml:"drop_over_limit"`
}

type server struct {
	Listen string `yaml:"listen"`
	Paths  []path `yaml:"paths"`
}

type Config struct {
	Server server `yaml:"server"`
}

func FromYaml(yamlReader io.Reader) (cfg Config, err error) {
	rawYaml, err := io.ReadAll(yamlReader)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(rawYaml, &cfg)
	return
}
