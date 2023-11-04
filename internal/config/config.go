package config

import (
	"gopkg.in/yaml.v3"
	"io"
)

// path defines properties of a certain URL path the reverse proxy will forward requests from.
type path struct {
	// Location is a URL path pattern that all incoming requests will be matched against. This pattern will be used
	// to register an HTTP handler which will perform the matching.
	Location string `yaml:"location"`

	// Target is the web URL of the origin server. It will be used to construct a full requesting URL to that server
	// when forwarding a request.
	Target string `yaml:"target"`

	// ConnectionLimit is a maximum number of concurrent requests to the origin server.
	ConnectionLimit int `yaml:"connection_limit"`

	// DropOverLimit describes whether the requests that exceed the ConnectionLimit have to be dropped or simply
	// delayed. If this property is true, limit-exceeding requests will be discarded and 503 status code returned.
	DropOverLimit bool `yaml:"drop_over_limit"`
}

// server defines configuration parameters for web server of the reverse proxy.
type server struct {
	// Listen is a port on which the reverse proxy is listening for incoming HTTP requests.
	Listen string `yaml:"listen"`

	// Paths stores all the
	Paths []path `yaml:"paths"`
}

// Config stores global reverse proxy configuration.
type Config struct {
	// Server stores web server related configuration.
	Server server `yaml:"server"`
}

// FromYaml parses YAML data from io.Reader into a Config.
func FromYaml(yamlReader io.Reader) (cfg Config, err error) {
	rawYaml, err := io.ReadAll(yamlReader)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(rawYaml, &cfg)
	return
}
