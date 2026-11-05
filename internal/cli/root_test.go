package cli

import (
	"bytes"
	"testing"
)

func TestRootCmdHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd := NewRootCmd(&out, &errOut)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("nt-inspect <codigo-nt>")) {
		t.Fatalf("Expected help message to contain usage syntax, got: %s", out.String())
	}
}

func TestRootCmdMissingArg(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd := NewRootCmd(&out, &errOut)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("Expected error when no argument provided, got nil")
	}
}
