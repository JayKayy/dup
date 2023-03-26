package duplicate

import (
	"crypto/sha256"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"

	"dup/pkg/config"
)

var (
	fileMap    = map[string][]string{}
	duplicates []string
	conf       *config.Config
)

func SetConfig(c *config.Config) {
	conf = c
}

func ProcessFiles(dir string) error {

	// open the directory
	fileList, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("opening directory %v\n", err)
	}

	for _, dirEntry := range fileList {
		path := fmt.Sprintf("%s/%s", dir, dirEntry.Name())
		metadata, err := os.Stat(path)
		if err != nil {
			log.Debugf("failed stating file, %v\n", err)
			continue
		}
		if metadata.IsDir() && !conf.Recurse {
			// Is a directory and we do not want to recurse
			log.Debugf("recurse=%t - Dir:%t skipping directory %v", conf.Recurse, metadata.IsDir(), path)
			continue
		} else if metadata.IsDir() && conf.Recurse {
			// Is a dir and we do want to recuse
			err := ProcessFiles(path)
			if err != nil {
				log.Debugf("recurse=%t - skipping directory %v", conf.Recurse, path)
			}
			continue
		} else {
			// Is not a directory but is a file to hash
			err = hashFiletoMap(path)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
func GetFileMap() map[string][]string {
	return fileMap
}
func GetDuplicates() []string {
	var result []string
	for _, v := range fileMap {
		if len(v) > 1 {
			for _, s := range v {
				result = append(result, s)
			}
		}
	}
	return result
}

func hashFiletoMap(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Reading file contents %v\n", err)
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Debugf("skipping %s. failed to hash file. %v\n", file.Name(), err)
		return err
	}
	// if the hash is in the map, add the file to the duplicates list
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	fileMap[hash] = append(fileMap[hash], path)

	return nil
}
