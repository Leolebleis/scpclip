package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Uploader interface {
	Upload(localPath, host, remotePath string) error
}

type SSHUploader struct {
	command func(name string, arg ...string) *exec.Cmd
}

func NewSSHUploader() *SSHUploader {
	return &SSHUploader{command: exec.Command}
}

func (u *SSHUploader) Upload(localPath, host, remotePath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	cmd := u.command("ssh", host, fmt.Sprintf("umask 077 && cat > %s", remotePath))
	cmd.Stdin = f
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	return nil
}
