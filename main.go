package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var version = "dev"

func run(host, dir string, cb Clipboard, up Uploader) error {
	data, err := cb.ReadImage()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("scpclip_%d.png", time.Now().Unix())

	// Stage the image in a uniquely-named local temp file. It must NOT reuse
	// the remote filename: when the host is local and --dir is the temp dir,
	// tmpPath would equal remotePath and the deferred cleanup below would
	// delete the file we just uploaded.
	tmpFile, err := os.CreateTemp("", "scpclip-*.png")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()          //nolint:errcheck // reopened via os.WriteFile below
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

	success("Uploaded to %s:%s", host, remotePath)
	success("Path copied to clipboard")
	fmt.Println(remotePath)
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `scpclip - clipboard image to SSH host in one command

Usage:
  scpclip [--host user@host] [--dir /remote/dir]
  scpclip default [host]    set or show default host
  scpclip update            update to latest version
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

	if len(os.Args) >= 2 && os.Args[1] == "update" {
		if err := doUpdate(); err != nil {
			fail("%s", err)
			os.Exit(1)
		}
		return
	}

	host := flag.String("host", "", "SSH host (overrides SCPCLIP_HOST env var)")
	dir := flag.String("dir", "", "Remote directory (overrides SCPCLIP_DIR env var, default /tmp)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("scpclip", version)
		if version != "dev" {
			if latest, err := checkLatestVersion(); err == nil && latest != version {
				fmt.Fprintln(os.Stderr)
				success("Update available: %s → %s", version, latest)
				hint("Run 'scpclip update' to update")
			}
		}
		return
	}

	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		fail("Could not read config: %v", cfgErr)
	}

	if *host == "" {
		*host = os.Getenv("SCPCLIP_HOST")
	}
	if *host == "" {
		*host = cfg.Host
	}
	if *host == "" {
		fail("No host specified")
		hint("Run 'scpclip default <host>' to set a default, or use --host")
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
		fail("ssh not found on PATH")
		os.Exit(1)
	}

	cb := &OSClipboard{}
	up := NewSSHUploader()

	if err := run(*host, *dir, cb, up); err != nil {
		fail("%s", err)
		if err.Error() == "no image in clipboard" {
			hint("Take a screenshot first (Win+Shift+S, Cmd+Shift+4, etc.)")
		}
		os.Exit(1)
	}
}

func handleDefault(args []string) {
	cfg, err := loadConfig()
	if err != nil {
		fail("%s", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		if cfg.Host == "" {
			fmt.Println("no default host set")
			hint("Run 'scpclip default <host>' to set one")
		} else {
			fmt.Println(cfg.Host)
		}
		return
	}

	cfg.Host = args[0]
	if err := saveConfig(cfg); err != nil {
		fail("%s", err)
		os.Exit(1)
	}
	success("Default host set to %s", bold(args[0]))
}
