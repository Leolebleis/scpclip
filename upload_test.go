package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)



func fakeExecCommand(t *testing.T, shouldFail bool) func(string, ...string) *exec.Cmd {
	t.Helper()
	return func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", name}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		env := []string{"GO_WANT_HELPER_PROCESS=1"}
		if shouldFail {
			env = append(env, "GO_HELPER_FAIL=1")
		}
		cmd.Env = env
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	if os.Getenv("GO_HELPER_FAIL") == "1" {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestUpload_Success(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(tmpFile, []byte("fake image"), 0600); err != nil {
		t.Fatal(err)
	}

	u := &SSHUploader{command: fakeExecCommand(t, false)}
	err := u.Upload(tmpFile, "user@host", "/tmp/test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpload_SSHFails(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(tmpFile, []byte("fake image"), 0600); err != nil {
		t.Fatal(err)
	}

	u := &SSHUploader{command: fakeExecCommand(t, true)}
	err := u.Upload(tmpFile, "user@host", "/tmp/test.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpload_BadLocalPath(t *testing.T) {
	u := &SSHUploader{command: fakeExecCommand(t, false)}
	err := u.Upload("/nonexistent/file.png", "user@host", "/tmp/test.png")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpload_CommandArgs(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(tmpFile, []byte("fake image"), 0600); err != nil {
		t.Fatal(err)
	}

	var capturedName string
	var capturedArgs []string
	u := &SSHUploader{
		command: func(name string, args ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = args
			return fakeExecCommand(t, false)(name, args...)
		},
	}

	_ = u.Upload(tmpFile, "user@host", "/tmp/test.png")

	if capturedName != "ssh" {
		t.Fatalf("expected command 'ssh', got %q", capturedName)
	}
	if capturedArgs[0] != "user@host" {
		t.Fatalf("expected host 'user@host', got %q", capturedArgs[0])
	}
	if !strings.Contains(capturedArgs[1], "umask 077") {
		t.Fatalf("expected 'umask 077' in remote command, got %q", capturedArgs[1])
	}
	if !strings.Contains(capturedArgs[1], "/tmp/test.png") {
		t.Fatalf("expected remote path in command, got %q", capturedArgs[1])
	}
}
