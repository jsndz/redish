package aof

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jsndz/redish/internal/config"
	"github.com/jsndz/redish/util"
)

type ManifestEntry struct {
	File string
	Seq  int
	Type string
}

type AOF struct {
	cfg           *config.Config
	mu            sync.Mutex
	timesAppended int
}

var writeCommands = map[string]bool{
	"SET":     true,
	"INCR":    true,
	"RPUSH":   true,
	"MULTI":   true,
	"EXEC":    true,
	"DISCARD": true,
}

func IsWriteCommand(cmd string) bool {
	return writeCommands[cmd]
}

func New(cfg *config.Config) (*AOF, error) {
	if !cfg.AppendOnly {
		return nil, nil
	}

	a := &AOF{cfg: cfg}

	folder := filepath.Join(cfg.Dir, cfg.AppendDirName)
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		return nil, err
	}

	entries, err := a.ParseManifest()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(entries) > 0 {
		a.timesAppended = entries[len(entries)-1].Seq
	} else {
		a.timesAppended = 1
		filePath := a.getAppendFilePath()

		file, err := os.OpenFile(
			filePath,
			os.O_CREATE|os.O_APPEND|os.O_RDWR,
			0644,
		)
		if err != nil {
			return nil, err
		}
		file.Close()

		manifestPath := a.getAppendManifestFilePath()
		manifest, err := os.OpenFile(
			manifestPath,
			os.O_CREATE|os.O_APPEND|os.O_RDWR,
			0644,
		)
		if err != nil {
			return nil, err
		}
		defer manifest.Close()

		a.writeToManifest(
			manifest,
			cfg.AppendFileName,
			a.timesAppended,
			"incr",
		)
	}

	return a, nil
}

func (a *AOF) writeToManifest(
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

func (a *AOF) getAppendFilePath() string {
	name := strings.TrimSuffix(a.cfg.AppendFileName, ".aof")

	// The original code used cfg.AppendType which was not initialized.
	// In SetConfig it used "incr".
	return filepath.Join(
		a.cfg.Dir,
		a.cfg.AppendDirName,
		fmt.Sprintf(
			"%s.%d.%s.aof",
			name,
			a.timesAppended,
			"incr",
		),
	)
}

func (a *AOF) getAppendManifestFilePath() string {
	name := strings.TrimSuffix(a.cfg.AppendFileName, ".aof")

	return filepath.Join(
		a.cfg.Dir,
		a.cfg.AppendDirName,
		fmt.Sprintf(
			"%s.manifest", name,
		),
	)
}

func (a *AOF) Write(data string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	filePath := a.getAppendFilePath()

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

	if a.cfg.AppendFsyncMode == "always" {
		file.Sync()
	}
}

func (a *AOF) ParseManifest() ([]ManifestEntry, error) {
	manifestPath := a.getAppendManifestFilePath()
	b, err := os.ReadFile(manifestPath)
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
				a.cfg.Dir,
				a.cfg.AppendDirName,
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

func (a *AOF) Restore() [][]interface{} {
	entries, err := a.ParseManifest()
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
				if ok && len(arr) > 0 {
					rawCommands = append(rawCommands, arr)
				}
				idx += n
			}
		}
	}

	return rawCommands
}
