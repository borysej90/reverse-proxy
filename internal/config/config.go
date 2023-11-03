package config

import (
	"gopkg.in/yaml.v3"
	"io"
)

type path struct {
	Location        string
	Target          string
	ConnectionLimit int
}

type server struct {
	Listen string
	Paths  []path
}

type Config struct {
	Server server
}

func FromYaml(yamlReader io.Reader) (cfg Config, err error) {
	rawYaml, err := io.ReadAll(yamlReader)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(rawYaml, &cfg)
	return
}
