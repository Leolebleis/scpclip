package main

import (
	"fmt"
	"os"
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

func main() {}
