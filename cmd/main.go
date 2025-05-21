package main

import (
	"dup/pkg/config"
	"dup/pkg/duplicate"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
)

func main() {
	var recurse, verbose, help bool
	var loglevel slog.LevelVar
	var logger *slog.Logger
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &loglevel,
	}))
	conf := config.Config{}

	flag.BoolVar(&recurse, "r", false, "recursively search directories beneath the specified directories.")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging.")
	flag.BoolVar(&help, "h", false, "display help message.")
	flag.Var(&conf, "d", "directories to search for duplicate files.")
	flag.Parse()

	if len(conf.Directories) == 0 || help {
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

	err := conf.ResolvePaths()
	if err != nil {
		slog.Error("resolving relative paths", "err", err)
		return
	}
	duplicate.SetConfig(&conf)

	var wg sync.WaitGroup
	var mut sync.Mutex
	for _, dir := range conf.Directories {
		dir := dir
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
	fmt.Print(ProcessFileMap())
}

func ProcessFileMap() string {
	fileMap := duplicate.GetHashMap()
	sb := strings.Builder{}

	for _, list := range fileMap {
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
