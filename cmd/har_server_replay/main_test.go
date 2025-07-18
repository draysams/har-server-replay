package main

import (
	"os"
	"strings"
	"testing"
)

func writeTempHAR(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.har")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestRun_MissingHarFile(t *testing.T) {
	err := Run([]string{})
	if err == nil || err.Error() != "--har-file is required" {
		t.Errorf("expected error for missing har-file, got %v", err)
	}
}

func TestRun_BadHarFile(t *testing.T) {
	badFile := writeTempHAR(t, "not json")
	err := Run([]string{"--har-file", badFile})
	if err == nil || !(strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "unexpected")) {
		t.Errorf("expected error for bad HAR file, got %v", err)
	}
}

func TestRun_BadPort(t *testing.T) {
	file := writeTempHAR(t, `{"log":{"entries":[]}}`)
	err := Run([]string{"--har-file", file, "--port", "notaport"})
	if err == nil {
		t.Errorf("expected error for bad port, got nil")
	}
}
