package config

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	Directories []string
	Recurse     bool
	IsTest      bool
	ReadOnyMode bool
}

// Search implements the flag.Value interface
func (c *Config) String() string {
	return fmt.Sprintf("%v", c.Directories)
}

func (c *Config) Set(val string) error {
	c.Directories = append(c.Directories, val)
	return nil
}

// Clean converts the directories to an absolute path.
func (c *Config) Clean() error {
	for i, v := range c.Directories {
		if v == "." {
			cwd, err := os.Getwd()
			if err != nil {
				slog.Error("getting working directory", "err", err)
				return err
			}
			c.Directories[i] = cwd
		}
	}
	return nil
}
