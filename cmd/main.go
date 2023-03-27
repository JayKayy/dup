package main

import (
	"dup/pkg/config"
	"dup/pkg/duplicate"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

func main() {

	var recurse, verbose, help bool
	var level log.Level
	conf := config.Config{}

	flag.BoolVar(&recurse, "r", false, "recursively search directories beneath the specified directories.")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging.")
	flag.BoolVar(&help, "h", false, "display help message.")
	flag.Var(&conf, "d", "a directory to search for duplicate files.")
	flag.Parse()

	if len(conf.Directories) == 0 || help {
		flag.PrintDefaults()
		return
	}

	if verbose {
		level = log.DebugLevel
	} else {
		level = log.InfoLevel
	}

	conf.Recurse = recurse
	conf.LogLevel = level

	//conf.IsTest = false
	//if conf.IsTest {
	//	conf.LogLevel = log.DebugLevel
	//	path, _ := filepath.Abs("./test")
	//	conf.Directories = append(conf.Directories, path)
	//}

	log.SetLevel(conf.LogLevel)
	err := conf.Clean()
	if err != nil {
		log.Fatalf("parsing a relative path: %v", err)
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
				log.Debugf("processing dir failed %v", err)
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
