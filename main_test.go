package main

import (
	"errors"
	"strings"
	"testing"
)

type mockClipboard struct {
	imageData []byte
	readErr   error
	writeErr  error
	written   string
}

func (m *mockClipboard) ReadImage() ([]byte, error) {
	return m.imageData, m.readErr
}

func (m *mockClipboard) WriteText(s string) error {
	m.written = s
	return m.writeErr
}

type mockUploader struct {
	err        error
	lastHost   string
	lastRemote string
}

func (m *mockUploader) Upload(localPath, host, remotePath string) error {
	m.lastHost = host
	m.lastRemote = remotePath
	return m.err
}

func TestRun_NoImage(t *testing.T) {
	cb := &mockClipboard{readErr: errors.New("no image in clipboard")}
	up := &mockUploader{}
	err := run("user@host", "/tmp", cb, up)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "no image in clipboard" {
		t.Fatalf("expected 'no image in clipboard', got %q", err.Error())
	}
}

func TestRun_Success(t *testing.T) {
	cb := &mockClipboard{imageData: []byte("fake png data")}
	up := &mockUploader{}
	err := run("user@host", "/tmp", cb, up)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if up.lastHost != "user@host" {
		t.Fatalf("expected host 'user@host', got %q", up.lastHost)
	}
	if cb.written == "" {
		t.Fatal("expected clipboard write, got empty")
	}
	if up.lastRemote != cb.written {
		t.Fatalf("remote path %q doesn't match clipboard %q", up.lastRemote, cb.written)
	}
}

func TestRun_UploadFails(t *testing.T) {
	cb := &mockClipboard{imageData: []byte("fake png data")}
	up := &mockUploader{err: errors.New("ssh: connection refused")}
	err := run("user@host", "/tmp", cb, up)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "upload failed") {
		t.Fatalf("expected 'upload failed' in error, got %q", err.Error())
	}
}

func TestRun_ClipboardWriteFails(t *testing.T) {
	cb := &mockClipboard{
		imageData: []byte("fake png data"),
		writeErr:  errors.New("clipboard write failed"),
	}
	up := &mockUploader{}
	err := run("user@host", "/tmp", cb, up)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "clipboard write failed") {
		t.Fatalf("expected 'clipboard write failed' in error, got %q", err.Error())
	}
}
