package main

import (
	"dup/pkg/config"
	"dup/pkg/duplicate"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

func main() {
	var recurse, verbose, hashes, help bool
	var dirs, output string
	var loglevel slog.LevelVar
	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &loglevel,
	}))

	conf := config.Config{}

	flag.BoolVar(&recurse, "r", false, "recursively search directories beneath the specified directories.")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging.")
	flag.BoolVar(&help, "h", false, "display help message.")
	flag.StringVar(&dirs, "d", ".", "directories to search for duplicate files.")
	flag.StringVar(&output, "o", "text", "display mode for results. options: \"text\", \"json\"")
	flag.BoolVar(&hashes, "x", false, "whether to change the resulting map's keys to the file hashes. Otherwise the first file processed with that hash is used as the key.")

	flag.Parse()

	if help {
		flag.PrintDefaults()
		return
	}
	err := conf.SetDirectories(dirs)
	if err != nil {
		slog.Error("setting directories", "err", err)
	}

	if len(conf.Directories) == 0 {
		slog.Error("no directories specified")
		flag.PrintDefaults()
		return
	}

	if verbose {
		loglevel.Set(slog.LevelDebug)
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: &loglevel,
		}))
	}
	slog.SetDefault(logger)

	conf.Recurse = recurse

	err = conf.ResolvePaths()
	if err != nil {
		slog.Error("resolving relative paths", "err", err)
		return
	}
	duplicate.SetConfig(&conf)

	var wg sync.WaitGroup
	var mut sync.Mutex
	for _, dir := range conf.Directories {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := duplicate.ProcessFiles(dir, &mut)
			if err != nil {
				slog.Debug("processing dir failed", "err", err)
			}
		}()
	}
	wg.Wait()
	switch output {
	case "text":
		fmt.Print(TextFileMap())

	case "json":
		fmt.Print(JSONFileMap(hashes))

	default:
		fmt.Print(TextFileMap())
	}
}

func JSONFileMap(hashes bool) string {
	fileMap := duplicate.GetHashMap().Map
	var marshalTarget map[string][]string
	if hashes {
		// use hashes as keys and all files as list
		marshalTarget = map[string][]string{}
		for hash, list := range fileMap {
			if len(list) > 1 {
				marshalTarget[hash] = list
			}
		}
	} else {
		// use the first file as a key and rest as list
		marshalTarget = map[string][]string{}
		for _, list := range fileMap {
			if len(list) > 1 {
				marshalTarget[list[0]] = list[1:]
			}
		}
	}
	js, err := json.Marshal(marshalTarget)
	if err != nil {
		slog.Error("marshalling filemap", "map", fileMap, "err", err)
	}
	return string(js)

}

func TextFileMap() string {
	fileMap := duplicate.GetHashMap()
	sb := strings.Builder{}

	for _, list := range fileMap.Map {
		if len(list) > 1 {
			for i, path := range list {
				if i == len(list)-1 {
					sb.WriteString(path)
				} else {
					sb.WriteString(path + " = ")
				}
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
