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

type SyncHashMap struct {
	Map map[string][]string
	Mut *sync.Mutex
}
type SyncPathMap struct {
	Map map[string]bool
	Mut *sync.Mutex
}

var (
	hashMap = SyncHashMap{
		Map: map[string][]string{},
		Mut: &sync.Mutex{},
	}
	pathMap = SyncPathMap{
		Map: map[string]bool{},
		Mut: &sync.Mutex{},
	}

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

	var wg sync.WaitGroup
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
				slog.Debug("skipping directory", "recurse", conf.Recurse, "dir", path)
				continue
			} else {
				// Is a dir and we want to recuse
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := ProcessFiles(path, mut); err != nil {
						slog.Error("processing files",
							"recurse", conf.Recurse,
							"skipping directory", path,
							"err", err)
					}
				}()
			}
		} else {
			// Is not a directory but is a file to hash
			_, ok := pathMap.Map[path]
			if ok {
				// file has already been processed before
				continue
			} else {
				pathMap.Mut.Lock()
				pathMap.Map[path] = true
				pathMap.Mut.Unlock()
			}
			if err := hashFiletoMap(path); err != nil {
				slog.Error("error hashing file", "path", path, "err", err)
			}
		}
	}
	wg.Wait()
	return nil
}

func GetHashMap() SyncHashMap {
	return hashMap
}

func GetAllDuplicates() []string {
	var result []string
	for _, v := range hashMap.Map {
		if len(v) > 1 {
			result = append(result, v...)
		}
	}
	return result
}

func hashFiletoMap(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reading file contents %v", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			slog.Error("closing file", "file", file)
		}
	}()

	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		slog.Error("skipping... failed to hash file", "file", file.Name(), "err", err)
		return err
	}

	// if the hash is in the map, add the file to the duplicates list
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	hashMap.Mut.Lock()
	hashMap.Map[hash] = append(hashMap.Map[hash], path)
	hashMap.Mut.Unlock()
	slog.Debug("processed file", "file", path, "hash", hash)

	return nil
}
