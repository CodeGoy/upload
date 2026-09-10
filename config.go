package main

import (
	"fmt"
	"log"
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
	for _, line := range strings.Split(string(bytes), "\n") {
		if len(line) == 0 {
			fmt.Println("empty line")
			continue
		}
		if line[0:1] == "#" {
			fmt.Println("line is a comment")
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

func testConfig() {
	config, err := loadConfig("./upload.conf")
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
		return
	}
	fmt.Printf("Config: %+v\n", config)
	//loadConfig("/etc/upload.conf")

}
