package duplicate

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
type DupeSearch struct {
	PathMap *SyncPathMap
	HashMap *SyncHashMap
	Config  *Config
}

func (d *DupeSearch) Init(c *Config) {
	d.Config = c
	d.HashMap = &SyncHashMap{
		Map: map[string][]string{},
		Mut: &sync.Mutex{},
	}
	d.PathMap = &SyncPathMap{
		Map: map[string]bool{},
		Mut: &sync.Mutex{},
	}
}

func (d *DupeSearch) GetHashMap() map[string][]string {
	return d.HashMap.Map
}

func (d *DupeSearch) Process(path string) error {
	metadata, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed stating file: %w", err)
	}

	if metadata.IsDir() {
		if err := d.ProcessDir(path); err != nil {
			return fmt.Errorf("processing dir %s: %w", path, err)
		}
	} else {
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", path, err)
		}

		if err := d.ProcessFile(file); err != nil {
			return fmt.Errorf("processing file %s: %w", path, err)
		}
		if err = file.Close(); err != nil {
			return fmt.Errorf("closing file %s: %w", path, err)
		}
	}

	return nil
}

func (d *DupeSearch) ProcessFile(file *os.File) error {
	metadata, err := os.Stat(file.Name())
	if err != nil {
		return fmt.Errorf("failed stating file: %w", err)
	}

	if metadata.IsDir() {
		return fmt.Errorf("file to process is a dir: %w", err)
	}

	path, err := filepath.Abs(file.Name())
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	_, ok := d.PathMap.Map[path]
	if ok {
		return nil
	} else {
		d.PathMap.Mut.Lock()
		d.PathMap.Map[path] = true
		d.PathMap.Mut.Unlock()
	}
	if err := d.hashFiletoMap(path); err != nil {
		slog.Error("hashing file", "path", path, "err", err)
		return fmt.Errorf("hashing file: %w", err)
	}

	return nil
}

func (d *DupeSearch) ProcessDir(dir string) error {
	metadata, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed stating file: %w", err)
	}

	if !metadata.IsDir() {
		return fmt.Errorf("file is not a dir: %w", err)
	}

	fileList, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("opening directory %w", err)
	}

	var wg sync.WaitGroup
	for _, dirEntry := range fileList {
		path := fmt.Sprintf("%s/%s", dir, dirEntry.Name())
		if !d.Config.Recurse {
			// Is a directory and we do NOT want to recurse
			slog.Debug("skipping directory", "recurse", d.Config.Recurse, "dir", path)
			continue
		} else {
			// Is a dir and we want to recuse
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.Process(path); err != nil {
					slog.Error("processing files",
						"recurse", d.Config.Recurse,
						"skipping directory", path,
						"err", err)
				}
			}()
		}
	}
	wg.Wait()
	return nil
}

func (d *DupeSearch) GetAllDuplicates() []string {
	var result []string
	d.HashMap.Mut.Lock()
	for _, v := range d.HashMap.Map {
		if len(v) > 1 {
			result = append(result, v...)
		}
	}
	d.HashMap.Mut.Unlock()
	return result
}

func (d *DupeSearch) hashFiletoMap(path string) error {
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
	d.HashMap.Mut.Lock()
	d.HashMap.Map[hash] = append(d.HashMap.Map[hash], path)
	d.HashMap.Mut.Unlock()
	slog.Debug("processed file", "file", path, "hash", hash)

	return nil
}
