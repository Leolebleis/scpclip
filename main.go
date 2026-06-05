package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var version = "dev"

func run(host, dir string, cb Clipboard, up Uploader) error {
	data, err := cb.ReadImage()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("scpclip_%d.png", time.Now().Unix())
	tmpPath := filepath.Join(os.TempDir(), filename)
	defer os.Remove(tmpPath)

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	remotePath := dir + "/" + filename
	if err := up.Upload(tmpPath, host, remotePath); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if err := cb.WriteText(remotePath); err != nil {
		return fmt.Errorf("clipboard write failed: %w", err)
	}

	fmt.Println(remotePath)
	return nil
}

func main() {
	host := flag.String("host", "", "SSH host (overrides SCPCLIP_HOST env var)")
	dir := flag.String("dir", "", "Remote directory (overrides SCPCLIP_DIR env var, default /tmp)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("scpclip", version)
		return
	}

	if *host == "" {
		*host = os.Getenv("SCPCLIP_HOST")
	}
	if *host == "" {
		fmt.Fprintln(os.Stderr, "error: no host specified (use --host or set SCPCLIP_HOST)")
		os.Exit(1)
	}

	if *dir == "" {
		*dir = os.Getenv("SCPCLIP_DIR")
	}
	if *dir == "" {
		*dir = "/tmp"
	}

	if _, err := exec.LookPath("ssh"); err != nil {
		fmt.Fprintln(os.Stderr, "error: ssh not found on PATH")
		os.Exit(1)
	}

	cb := &OSClipboard{}
	up := NewSSHUploader()

	if err := run(*host, *dir, cb, up); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
