package main

import (
	"dup/pkg/config"
	"dup/pkg/duplicate"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	testDir = "./test"
)

func main() {

	conf := config.Config{
		Directories: []string{},
		Recurse:     false,
		IsTest:      true,
		ReadOnyMode: true,
		LogLevel:    log.InfoLevel,
	}

	if conf.IsTest {
		conf.LogLevel = log.DebugLevel
		//	path, _ := filepath.Abs(testDir)
		//conf.Directories = append(conf.Directories, path)
	}
	log.SetLevel(conf.LogLevel)

	flag.Var(&conf, "d", "a directory to search for duplicate files.")
	flag.Parse()

	log.Debug(conf.Directories)
	conf.Clean()
	log.Debug(conf.Directories)

	duplicate.SetConfig(&conf)

	for _, dir := range conf.Directories {

		err := duplicate.ProcessFiles(dir)
		if err != nil {
			log.Errorf("%v\n", err)
		}

	}

	fmt.Println(ProcessFileMap())
}

func ProcessFileMap() string {
	fileMap := duplicate.GetFileMap()
	sb := strings.Builder{}

	for _, list := range fileMap {
		if len(list) > 1 {
			for _, path := range list {
				sb.WriteString(path + " ")
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
