package config

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFromYaml(t *testing.T) {
	var tests = []struct {
		name     string
		yaml     string
		expected Config
	}{
		{
			name:     "empty YAML",
			yaml:     "",
			expected: Config{},
		},
		{
			name: "happy pass",
			yaml: `
server:
  listen: 8080
  paths:
    - location: /
      target: http://127.0.0.1:8081
      connection_limit: 100
      drop_over_limit: true
`,
			expected: Config{Server: server{
				Listen: "8080",
				Paths: []path{
					{
						Location:        "/",
						Target:          "http://127.0.0.1:8081",
						ConnectionLimit: 100,
						DropOverLimit:   true,
					},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.yaml)
			actual, err := FromYaml(reader)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
