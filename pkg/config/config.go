package config

import (
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Directories []string
	Recurse     bool
	IsTest      bool
	ReadOnyMode bool
	LogLevel    log.Level
}
