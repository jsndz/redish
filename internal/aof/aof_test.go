package aof

import (
	"os"
	"testing"

	"github.com/jsndz/redish/internal/config"
)

func TestAOF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aof_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		Dir:             tempDir,
		AppendOnly:      true,
		AppendDirName:   "appendonlydir",
		AppendFileName:  "appendonly.aof",
		AppendFsyncMode: "everysec",
	}

	// Test New
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create AOF: %v", err)
	}
	if a == nil {
		t.Fatal("Expected AOF instance, got nil")
	}

	// Test Write
	cmd1 := "*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n"
	a.Write(cmd1)

	cmd2 := "*2\r\n$4\r\nINCR\r\n$1\r\na\r\n"
	a.Write(cmd2)

	// Test Restore
	commands := a.Restore()
	if len(commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(commands))
	}

	if commands[0][0] != "SET" {
		t.Errorf("Expected first command SET, got %v", commands[0][0])
	}
	if commands[1][0] != "INCR" {
		t.Errorf("Expected second command INCR, got %v", commands[1][0])
	}

	// Test ParseManifest
	entries, err := a.ParseManifest()
	if err != nil {
		t.Fatalf("Failed to parse manifest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 manifest entry, got %d", len(entries))
	}
	if entries[0].Seq != 1 {
		t.Errorf("Expected seq 1, got %d", entries[0].Seq)
	}
}

func TestIsWriteCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"SET", true},
		{"GET", false},
		{"INCR", true},
		{"PING", false},
		{"RPUSH", true},
		{"MULTI", true},
		{"EXEC", true},
		{"DISCARD", true},
	}

	for _, tt := range tests {
		if got := IsWriteCommand(tt.cmd); got != tt.want {
			t.Errorf("IsWriteCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
		}
	}
}
