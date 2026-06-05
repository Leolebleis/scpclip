package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Clipboard interface {
	ReadImage() ([]byte, error)
	WriteText(s string) error
}

type OSClipboard struct{}

func (c *OSClipboard) ReadImage() ([]byte, error) {
	switch runtime.GOOS {
	case "darwin":
		return c.readDarwin()
	case "linux":
		return c.readLinux()
	case "windows":
		return c.readWindows()
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (c *OSClipboard) readDarwin() ([]byte, error) {
	if _, err := exec.LookPath("pngpaste"); err != nil {
		return nil, errors.New("pngpaste not found (install with: brew install pngpaste)")
	}
	out, err := exec.Command("pngpaste", "-").Output()
	if err != nil {
		return nil, errors.New("no image in clipboard")
	}
	return out, nil
}

func (c *OSClipboard) readLinux() ([]byte, error) {
	if path, err := exec.LookPath("xclip"); err == nil {
		out, err := exec.Command(path, "-selection", "clipboard", "-target", "image/png", "-o").Output()
		if err != nil || len(out) == 0 {
			return nil, errors.New("no image in clipboard")
		}
		return out, nil
	}
	if path, err := exec.LookPath("wl-paste"); err == nil {
		out, err := exec.Command(path, "--type", "image/png").Output()
		if err != nil || len(out) == 0 {
			return nil, errors.New("no image in clipboard")
		}
		return out, nil
	}
	return nil, errors.New("no clipboard tool found (install xclip or wl-clipboard)")
}

func (c *OSClipboard) readWindows() ([]byte, error) {
	tmpFile := filepath.Join(os.TempDir(), "scpclip_read.png")
	defer os.Remove(tmpFile) //nolint:errcheck // best-effort cleanup of temp file

	script := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms
$img = [System.Windows.Forms.Clipboard]::GetImage()
if ($null -eq $img) { exit 1 }
$img.Save('%s')`, strings.ReplaceAll(tmpFile, `\`, `\\`))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if err := cmd.Run(); err != nil {
		return nil, errors.New("no image in clipboard")
	}
	return os.ReadFile(tmpFile)
}

func (c *OSClipboard) WriteText(s string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(s)
		return cmd.Run()
	case "linux":
		return c.writeLinux(s)
	case "windows":
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
			fmt.Sprintf("Set-Clipboard -Value '%s'", strings.ReplaceAll(s, "'", "''")))
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (c *OSClipboard) writeLinux(s string) error {
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(s)
		return cmd.Run()
	}
	if path, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(path)
		cmd.Stdin = strings.NewReader(s)
		return cmd.Run()
	}
	return errors.New("no clipboard tool found (install xclip or wl-clipboard)")
}
