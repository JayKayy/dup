package duplicate

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Directories []string
	Recurse     bool
}

// Implements the flag.Value interface
func (c *Config) String() string {
	return fmt.Sprintf("%v", c.Directories)
}

// SetDirectories takes a comma delimeted string of directories
// to be registered in the config.
func (c *Config) SetDirectories(val string) error {
	if c.Directories == nil {
		c.Directories = []string{}
	}
	dirs := strings.Split(val, ",")
	c.Directories = append(c.Directories, dirs...)
	return nil
}

// ResolvePaths converts the directories to an absolute path.
func (c *Config) ResolvePaths() error {
	if c.Directories == nil {
		return errors.New("directories is not instantiated")
	}
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
