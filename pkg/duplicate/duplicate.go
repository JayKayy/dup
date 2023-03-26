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
	fileMap    = map[string]string{}
	duplicates []string
	conf       *config.Config
)

func SetConfig(c *config.Config) {
	conf = c
}

func FindDuplicates(dir string) ([]string, error) {

	// open the directory
	fileList, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("opening directory %v\n", err)
	}

	for _, dirEntry := range fileList {
		path := fmt.Sprintf("%s/%s", dir, dirEntry.Name())
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("Reading file %v\n", err)
		}
		// hash the file
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			log.Debugf("skipping %s. failed to hash file. %v\n", err)
			file.Close()
			continue
		}
		// if the hash is in the map, add the file to the duplicates list
		hash := fmt.Sprintf("%x", hasher.Sum(nil))
		if _, ok := fileMap[hash]; ok {
			duplicates = append(duplicates, path)
		} else {
			fileMap[hash] = path
		}
		file.Close()
	}

	return duplicates, nil
}
