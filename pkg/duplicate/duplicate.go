package duplicate

import (
	"crypto/sha256"
	"dup/pkg/config"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/exp/slog"
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
		return fmt.Errorf("opening directory %v", err)
	}

	for _, dirEntry := range fileList {
		path := fmt.Sprintf("%s/%s", dir, dirEntry.Name())
		metadata, err := os.Stat(path)
		if err != nil {
			slog.Error("failed stating file", "err", err)
			continue
		}
		if metadata.IsDir() {
			if !conf.Recurse {
				// Is a directory and we do NOT want to recurse
				slog.Info("recurse=false, skipping directory", "dir", path)
				continue
			} else {
				// Is a dir and we want to recuse
				if err := ProcessFiles(path, mut); err != nil {
					slog.Error("processing files",
						"recurse", conf.Recurse,
						"skipping directory", path,
						"err", err)
				}
			}
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
			if err := hashFiletoMap(path, mut); err != nil {
				slog.Error("error hashing file", "path", path, "err", err)
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
			result = append(result, v...)
		}
	}
	return result
}

func hashFiletoMap(path string, mut *sync.Mutex) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reading file contents %v", err)
	}
	defer file.Close() //nolint: errcheck

	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		slog.Error("skipping... failed to hash file", "file", file.Name(), "err", err)
		return err
	}

	// if the hash is in the map, add the file to the duplicates list
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	mut.Lock()
	hashMap[hash] = append(hashMap[hash], path)
	mut.Unlock()
	slog.Debug("processed file", "file", path, "hash", hash)

	return nil
}
