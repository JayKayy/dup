package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	Directories []string
	Recurse     bool
	IsTest      bool
	ReadOnyMode bool
	LogLevel    log.Level
}

// Search implements the flag.Value interface
func (c *Config) String() string {
	return fmt.Sprintf("%v", c.Directories)
}

func (c *Config) Set(val string) error {
	c.Directories = append(c.Directories, val)
	return nil
}

func (c *Config) Clean() error {
	for i, v := range c.Directories {
		if v == "." {
			cwd, err := os.Getwd()
			if err != nil {
				log.Debugf("getting working directory, %v", err)
				return err
			}
			c.Directories[i] = cwd
		}
	}
	return nil
}
