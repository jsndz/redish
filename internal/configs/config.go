// configs/config.go
package configs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
type ManifestEntry struct {
	File string
	Seq  int
	Type string
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

	folder := filepath.Join(cfg.Dir, cfg.AppendDirName)

	err := os.MkdirAll(folder, 0755)
	if err != nil {
		panic(err)
	}
	entries, _ := cfg.ParseManifest()

	if len(entries) > 0 {
		cfg.TimesAppended = entries[len(entries)-1].Seq
	} else {
		cfg.TimesAppended = 1
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
			"incr",
		)
	}
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

	_, err = file.WriteString(data)
	if err != nil {
		panic(err)
	}

	if cfg.AppendFsyncMode == "always" {
		file.Sync()
	}
}

func (cfg *Config) ParseManifest() ([]ManifestEntry, error) {
	if !cfg.AppendOnly {
		return nil, nil
	}
	b, err := os.ReadFile(cfg.GetAppendManifestFilePath())
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	var entries []ManifestEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)

		if len(parts) < 6 ||
			parts[0] != "file" ||
			parts[2] != "seq" ||
			parts[4] != "type" {
			continue
		}
		seq, err := strconv.Atoi(parts[3])
		if err != nil {
			continue
		}

		entry := ManifestEntry{
			File: filepath.Join(
				cfg.Dir,
				cfg.AppendDirName,
				parts[1],
			),
			Seq:  seq,
			Type: parts[5],
		}
		entries = append(entries, entry)

	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Seq < entries[j].Seq
	})
	return entries, nil
}

func (cfg *Config) RestoreFromAppendFile() [][]interface{} {
	entries, err := cfg.ParseManifest()
	if err != nil {
		return nil
	}
	var rawCommands [][]interface{}
	for _, entry := range entries {
		switch entry.Type {
		case "b":
			continue
		case "i":
			data, err := os.ReadFile(entry.File)
			if err != nil {
				continue
			}

			content := string(data)
			idx := 0
			for idx < len(content) {
				cmd, n := util.RESPFormatter(content[idx:])
				arr, ok := cmd.([]interface{})
				if !ok && len(arr) > 0 {
					rawCommands = append(rawCommands, arr)
				}
				idx += n
			}
		}
	}

	return rawCommands
}
