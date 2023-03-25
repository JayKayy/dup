package main

import (
	"dup/pkg/config"
	"dup/pkg/duplicate"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const (
	testDir = "./test"
)

func main() {

	// define a string flag that takes a directory path
	// and has a default value of the current directory
	//	flag.StringVar(&directory, "dir", ".", "the directory to search")
	//	flag.Parse()

	conf := config.Config{
		Directories: nil,
		Recurse:     false,
		IsTest:      true,
		ReadOnyMode: true,
		LogLevel:    log.InfoLevel,
	}

	if conf.IsTest {
		conf.LogLevel = log.DebugLevel
		path, err := filepath.Abs(testDir)
		if err != nil {
			log.Fatalf("opening test dir %v", err)
		}
		conf.Directories = append(conf.Directories, path)
	}

	duplicate.SetConfig(&conf)

	for _, dir := range conf.Directories {
		metadata, err := os.Stat(dir)
		if err != nil {
			log.Warningf("failed stating duplicate, %v\n", err)
			continue
		}
		if !metadata.IsDir() {
			log.Warningf("skipping %v, not a directory", dir)
			continue
		}
		dupes, err := duplicate.FindDuplicates(dir)
		if err != nil {
			log.Errorf("%v\n", err)
		}
		if len(dupes) == 0 {
			log.Infof("no duplicate files found in %s\n", dir)
			return
		}

		fmt.Println("Duplicates found:")
		for _, dup := range dupes {
			log.Infof(dup)
		}
	}

}
