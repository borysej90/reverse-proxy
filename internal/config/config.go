package config

import (
	"gopkg.in/yaml.v3"
	"io"
)

// path defines properties of a particular URL path the reverse proxy will forward requests from.
type path struct {
	// Location is a URL path pattern that the proxy will match all incoming requests against. It will use this pattern
	// to register an HTTP handler, which will perform the matching.
	Location string `yaml:"location"`

	// Target is the web URL of the origin server. It will be used to construct a full requesting URL to that server
	// when forwarding a request.
	Target string `yaml:"target"`

	// ConnectionLimit is the maximum number of concurrent requests to the origin server.
	ConnectionLimit int `yaml:"connection_limit"`

	// DropOverLimit describes whether the requests that exceed the ConnectionLimit must be dropped or delayed. If this
	// property is true, limit-exceeding requests will be discarded, and the 503 status code will be returned.
	DropOverLimit bool `yaml:"drop_over_limit"`
}

// server defines configuration parameters for the web server of the reverse proxy.
type server struct {
	// Listen is a port on which the reverse proxy listens for incoming HTTP requests.
	Listen string `yaml:"listen"`

	// Paths stores the configuration of all forwarding paths and their targets.
	Paths []path `yaml:"paths"`
}

// Config stores global reverse proxy configuration.
type Config struct {
	// Server stores web server-related configuration.
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
