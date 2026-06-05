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
	defer os.Remove(tmpPath) //nolint:errcheck // best-effort cleanup of temp file

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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `scpclip - clipboard image to SSH host in one command

Usage:
  scpclip [--host user@host] [--dir /remote/dir]
  scpclip default [host]    set or show default host
  scpclip --version

Options:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Set a default host to skip --host every time:
  scpclip default pi
`)
	}

	if len(os.Args) >= 2 && os.Args[1] == "default" {
		handleDefault(os.Args[2:])
		return
	}

	host := flag.String("host", "", "SSH host (overrides SCPCLIP_HOST env var)")
	dir := flag.String("dir", "", "Remote directory (overrides SCPCLIP_DIR env var, default /tmp)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("scpclip", version)
		return
	}

	cfg, _ := loadConfig()

	if *host == "" {
		*host = os.Getenv("SCPCLIP_HOST")
	}
	if *host == "" {
		*host = cfg.Host
	}
	if *host == "" {
		fmt.Fprintln(os.Stderr, "error: no host specified (use --host, set SCPCLIP_HOST, or run: scpclip default <host>)")
		os.Exit(1)
	}

	if *dir == "" {
		*dir = os.Getenv("SCPCLIP_DIR")
	}
	if *dir == "" {
		*dir = cfg.Dir
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

func handleDefault(args []string) {
	if len(args) == 0 {
		cfg, err := loadConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if cfg.Host == "" {
			fmt.Println("no default host set")
		} else {
			fmt.Println(cfg.Host)
		}
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	cfg.Host = args[0]
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Printf("default host set to %s\n", args[0])
}
