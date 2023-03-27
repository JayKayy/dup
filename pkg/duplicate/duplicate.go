package duplicate

import (
	"crypto/sha256"
	"dup/pkg/config"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

var (
	hashMap = map[string][]string{}
	pathMap = map[string]bool{}

	conf *config.Config
)

func SetConfig(c *config.Config) {
	conf = c
}

func ProcessFiles(dir string, mut *sync.Mutex) error {

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
			log.Debugf("skipping directory %v", path)
			continue
		} else if metadata.IsDir() && conf.Recurse {
			// Is a dir and we do want to recuse
			err := ProcessFiles(path, mut)
			if err != nil {
				log.Debugf("recurse=%t - skipping directory %v", conf.Recurse, path)
			}
			continue
		} else {
			// Is not a directory but is a file to hash
			mut.Lock()
			_, ok := pathMap[path]
			mut.Unlock()

			if ok {
				// file has already been processed before
				continue
			} else {
				pathMap[path] = true
			}
			err = hashFiletoMap(path, mut)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func GetHashMap() map[string][]string {
	return hashMap
}

func GetAllDuplicates() []string {
	var result []string
	for _, v := range hashMap {
		if len(v) > 1 {
			for _, s := range v {
				result = append(result, s)
			}
		}
	}
	return result
}

func hashFiletoMap(path string, mut *sync.Mutex) error {

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Reading file contents %v\n", err)
	}
	defer file.Close()

	hasher := sha256.New()

	_, err = io.Copy(hasher, file)

	if err != nil {
		log.Debugf("skipping %s. failed to hash file. %v\n", file.Name(), err)
		return err
	}

	// if the hash is in the map, add the file to the duplicates list
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	mut.Lock()
	hashMap[hash] = append(hashMap[hash], path)
	mut.Unlock()
	log.Debugf("processed %s", path)

	return nil
}
