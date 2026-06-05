package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"runtime"
	"strings"
	"testing"
)

func TestExtractFromTarGz(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte("fake binary content")
	hdr := &tar.Header{Name: "scpclip", Size: int64(len(content)), Mode: 0o755}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := extractFromTarGz(buf.Bytes(), "scpclip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := new(bytes.Buffer)
	if _, err := result.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if result.String() != "fake binary content" {
		t.Fatalf("expected 'fake binary content', got %q", result.String())
	}
}

func TestExtractFromTarGz_NotFound(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := extractFromTarGz(buf.Bytes(), "scpclip")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtractFromZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.Create("scpclip.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("fake binary content")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := extractFromZip(buf.Bytes(), "scpclip.exe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := new(bytes.Buffer)
	if _, err := result.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if result.String() != "fake binary content" {
		t.Fatalf("expected 'fake binary content', got %q", result.String())
	}
}

func TestExtractFromZip_NotFound(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := extractFromZip(buf.Bytes(), "scpclip.exe")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindAsset(t *testing.T) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	release := githubRelease{
		TagName: "v0.2.0",
		Assets: []githubAsset{
			{Name: "scpclip_0.2.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux"},
			{Name: "scpclip_0.2.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin"},
			{Name: "scpclip_0.2.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/windows"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	asset, found := findAsset(release)
	if !found {
		t.Fatalf("expected to find asset for %s/%s", goos, goarch)
	}

	expectedSuffix := "_" + goos + "_" + goarch + ext
	if !strings.HasSuffix(asset.Name, expectedSuffix) {
		t.Fatalf("expected asset ending with %q, got %q", expectedSuffix, asset.Name)
	}
}

func TestFindAsset_NotFound(t *testing.T) {
	release := githubRelease{
		TagName: "v0.2.0",
		Assets: []githubAsset{
			{Name: "scpclip_0.2.0_freebsd_riscv64.tar.gz"},
		},
	}

	_, found := findAsset(release)
	if found {
		t.Fatal("expected no match for current OS/arch")
	}
}
