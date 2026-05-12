// configs/config.go
package configs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jsndz/redish/util"
)

type Config struct {
	Dir             string
	AppendOnly      bool
	AppendDirName   string
	AppendFileName  string
	AppendFsyncMode string
	TimesAppended   int
	AppendType      string

	mu sync.Mutex
}

func SetConfig() *Config {
	var cfg Config
	flag.StringVar(&cfg.Dir, "dir", ".", "data directory")

	flag.BoolVar(
		&cfg.AppendOnly,
		"appendonly",
		false,
		"enable append only mode",
	)

	flag.StringVar(
		&cfg.AppendDirName,
		"appenddirname",
		"appendonlydir",
		"append only directory name",
	)

	flag.StringVar(
		&cfg.AppendFileName,
		"appendfilename",
		"appendonly.aof",
		"append only file name",
	)

	flag.StringVar(
		&cfg.AppendFsyncMode,
		"appendfsync",
		"everysec",
		"fsync policy",
	)
	flag.Parse()

	if !cfg.AppendOnly {
		return &cfg
	}

	cfg.TimesAppended++

	folder := filepath.Join(cfg.Dir, cfg.AppendDirName)

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		panic(err)
	}

	filePath := cfg.GetAppendFilePath()

	file, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		panic(err)
	}
	file.Close()

	manifestPath := cfg.GetAppendManifestFilePath()
	manifest, err := os.OpenFile(
		manifestPath,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer manifest.Close()

	WriteToManifest(
		manifest,
		cfg.AppendFileName,
		cfg.TimesAppended,
		cfg.AppendType,
	)
	return &cfg
}

func WriteToManifest(
	file *os.File,
	filename string,
	seq int,
	appendType string,
) {
	t := "i"

	_, err := file.Write([]byte(
		fmt.Sprintf(
			"file %s.%d.%s.aof seq %d type %s\n",
			strings.TrimSuffix(filename, ".aof"),
			seq,
			appendType,
			seq,
			t,
		),
	))
	if err != nil {
		panic(err)
	}
}

func (cfg *Config) GetAppendFilePath() string {
	name := strings.TrimSuffix(cfg.AppendFileName, ".aof")

	return filepath.Join(
		cfg.Dir,
		cfg.AppendDirName,
		fmt.Sprintf(
			"%s.%d.%s.aof",
			name,
			cfg.TimesAppended,
			cfg.AppendType,
		),
	)
}

func (cfg *Config) GetAppendManifestFilePath() string {
	name := strings.TrimSuffix(cfg.AppendFileName, ".aof")

	return filepath.Join(
		cfg.Dir,
		cfg.AppendDirName,
		fmt.Sprintf(
			"%s.manifest", name,
		),
	)
}

func (cfg *Config) WriteToAppendFile(data string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	filePath := cfg.GetAppendFilePath()

	file, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(data + "\n")
	if err != nil {
		panic(err)
	}

	if cfg.AppendFsyncMode == "always" {
		file.Sync()
	}
}

func (cfg *Config) RestoreFromAppendFile() [][]interface{} {
	if !cfg.AppendOnly {
		return nil
	}
	b, err := os.ReadFile(cfg.GetAppendManifestFilePath())
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	var incrFile string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) < 6 {
			continue
		}

		if parts[5] != "i" {
			continue
		}
		incrFile = filepath.Join(cfg.Dir, cfg.AppendDirName, parts[1])
		break
	}
	data, err := os.ReadFile(incrFile)
	if err != nil {
		return nil
	}
	cmds := strings.Split(string(data), "\n")
	var rawCommands [][]interface{}
	for _, cmd := range cmds {
		command, _ := util.RESPFormatter(cmd)
		arr, ok := command.([]interface{})
		if !ok || len(arr) == 0 {
			continue
		}
		rawCommands = append(rawCommands, arr)
	}
	return rawCommands
}
