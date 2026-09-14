package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Endpoint string `ini:"endpoint"`
	Path     string `ini:"path"`
	TlsPort  string `ini:"tlsport"`
	Port     string `ini:"port"`
	Cert     string `ini:"cert"`
	Key      string `ini:"key"`
}

func loadConfig(configFile string) (*Config, error) {
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %s", err)
	}
	c := &Config{}
	for line := range strings.SplitSeq(string(bytes), "\n") {
		if len(line) == 0 {
			continue
		}
		if line[0:1] == "#" {
			continue
		}
		split := strings.Split(line, "=")
		if len(split) == 1 {
			continue
		}
		switch split[0] {
		case "endpoint":
			c.Endpoint = split[1]
		case "path":
			c.Path = split[1]
		case "tlsport":
			c.TlsPort = split[1]
		case "port":
			c.Port = split[1]
		case "cert":
			c.Cert = split[1]
		case "key":
			c.Key = split[1]
		}
	}
	return c, nil
}
