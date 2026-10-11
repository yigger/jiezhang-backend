package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExistingReportCannotTriggerApplyOrOverwriteInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.csv")
	content := []byte("synthetic source; must not be changed")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"--file", path, "--user-id", "1", "--book-id", "2", "--report", path, "--apply"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected rejection before connecting to database: %d", code)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, content) {
		t.Fatal("overwrote input")
	}
}

func TestCLIRejectsMissingIdentityAndUnknownSource(t *testing.T) {
	for _, args := range [][]string{{"--file", "example.csv"}, {"--file", "example.csv", "--user-id", "1", "--book-id", "2", "--source", "other"}} {
		var out, errOut bytes.Buffer
		if code := run(context.Background(), args, &out, &errOut); code != 2 {
			t.Fatalf("invalid CLI accepted: %d", code)
		}
	}
}
