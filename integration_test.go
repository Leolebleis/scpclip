//go:build integration

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var testBinary string

func TestMain(m *testing.M) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "1" {
		os.Exit(0)
		return
	}

	name := "scpclip"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dir, err := os.MkdirTemp("", "scpclip-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "creating temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir) //nolint:errcheck

	testBinary = filepath.Join(dir, name)
	cmd := exec.Command("go", "build", "-o", testBinary, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "building binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func buildBinary(t *testing.T) string {
	t.Helper()
	if testBinary == "" {
		t.Fatal("testBinary not set — TestMain didn't run")
	}
	return testBinary
}

type result struct {
	stdout   string
	stderr   string
	exitCode int
}

func runBinary(t *testing.T, bin string, args []string, env ...string) result {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("running binary: %v", err)
		}
	}
	return result{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: code,
	}
}

func testPNGPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs("testdata/test.png")
	if err != nil {
		t.Fatalf("resolving test.png path: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("test.png not found at %s: %v", path, err)
	}
	return path
}

func putImageOnClipboard(t *testing.T, pngPath string) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("xclip", "-selection", "clipboard", "-target", "image/png", "-i")
		f, err := os.Open(pngPath)
		if err != nil {
			t.Fatalf("opening test png: %v", err)
		}
		defer f.Close()
		cmd.Stdin = f
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("xclip set image: %v: %s", err, out)
		}
	case "darwin":
		script := fmt.Sprintf(`set the clipboard to (read POSIX file "%s" as «class PNGf»)`, pngPath)
		if out, err := exec.Command("osascript", "-e", script).CombinedOutput(); err != nil {
			t.Fatalf("osascript set image: %v: %s", err, out)
		}
	case "windows":
		script := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; Add-Type -AssemblyName System.Drawing; [System.Windows.Forms.Clipboard]::SetImage([System.Drawing.Image]::FromFile('%s'))`,
			strings.ReplaceAll(pngPath, `\`, `\\`))
		if out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput(); err != nil {
			t.Fatalf("powershell set image: %v: %s", err, out)
		}
	}
}

func clearClipboard(t *testing.T) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("xclip", "-selection", "clipboard", "-target", "image/png", "-i", "/dev/null")
		if err := cmd.Run(); err != nil {
			t.Logf("warning: clearing image clipboard: %v", err)
		}
		cmd2 := exec.Command("xclip", "-selection", "clipboard")
		cmd2.Stdin = strings.NewReader("cleared")
		if err := cmd2.Run(); err != nil {
			t.Logf("warning: clearing text clipboard: %v", err)
		}
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader("cleared")
		if err := cmd.Run(); err != nil {
			t.Logf("warning: clearing clipboard: %v", err)
		}
	case "windows":
		script := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Clipboard]::Clear()`
		if err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Run(); err != nil {
			t.Logf("warning: clearing clipboard: %v", err)
		}
	}
}

func readClipboardText(t *testing.T) string {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	case "darwin":
		out, err := exec.Command("pbpaste").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "Get-Clipboard").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	return ""
}

func sshAvailable(t *testing.T) bool {
	t.Helper()
	err := exec.Command("ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=2", "localhost", "echo", "ok").Run()
	return err == nil
}

func TestIntegration_Version(t *testing.T) {
	bin := buildBinary(t)
	r := runBinary(t, bin, []string{"--version"})
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "scpclip dev") {
		t.Fatalf("expected 'scpclip dev' in stdout, got %q", r.stdout)
	}
}

func TestIntegration_Help(t *testing.T) {
	bin := buildBinary(t)
	r := runBinary(t, bin, []string{"--help"})
	if !strings.Contains(r.stderr, "scpclip - clipboard image to SSH host in one command") {
		t.Fatalf("expected usage header in stderr, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "scpclip default [host]") {
		t.Fatalf("expected 'default' subcommand in usage, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "scpclip update") {
		t.Fatalf("expected 'update' subcommand in usage, got %q", r.stderr)
	}
}

func TestIntegration_NoHost(t *testing.T) {
	bin := buildBinary(t)
	r := runBinary(t, bin, nil,
		"SCPCLIP_HOST=",
		"HOME="+t.TempDir(),
		"XDG_CONFIG_HOME="+t.TempDir(),
		"AppData="+t.TempDir(),
	)
	if r.exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", r.exitCode)
	}
	if !strings.Contains(r.stderr, "No host specified") {
		t.Fatalf("expected 'No host specified' in stderr, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "scpclip default") {
		t.Fatalf("expected hint about 'scpclip default' in stderr, got %q", r.stderr)
	}
}

func TestIntegration_NoImage(t *testing.T) {
	bin := buildBinary(t)
	clearClipboard(t)
	r := runBinary(t, bin, []string{"--host", "localhost"})
	if r.exitCode != 1 {
		t.Fatalf("expected exit 1, got %d", r.exitCode)
	}
	if !strings.Contains(r.stderr, "no image in clipboard") {
		t.Fatalf("expected 'no image in clipboard' in stderr, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "screenshot") {
		t.Fatalf("expected screenshot hint in stderr, got %q", r.stderr)
	}
}

func TestIntegration_DefaultSet(t *testing.T) {
	bin := buildBinary(t)
	tmpHome := t.TempDir()
	env := []string{
		"XDG_CONFIG_HOME=" + tmpHome,
		"AppData=" + tmpHome,
		"HOME=" + tmpHome,
	}

	r := runBinary(t, bin, []string{"default", "testhost"}, env...)
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stderr, "Default host set to") {
		t.Fatalf("expected success message in stderr, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "testhost") {
		t.Fatalf("expected 'testhost' in stderr, got %q", r.stderr)
	}

	r = runBinary(t, bin, []string{"default"}, env...)
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "testhost") {
		t.Fatalf("expected 'testhost' in stdout, got %q", r.stdout)
	}
}

func TestIntegration_DefaultEmpty(t *testing.T) {
	bin := buildBinary(t)
	tmpHome := t.TempDir()
	env := []string{
		"XDG_CONFIG_HOME=" + tmpHome,
		"AppData=" + tmpHome,
		"HOME=" + tmpHome,
	}
	r := runBinary(t, bin, []string{"default"}, env...)
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "no default host set") {
		t.Fatalf("expected 'no default host set' in stdout, got %q", r.stdout)
	}
}

func TestIntegration_NoColor(t *testing.T) {
	bin := buildBinary(t)
	r := runBinary(t, bin, nil,
		"NO_COLOR=1",
		"SCPCLIP_HOST=",
		"HOME="+t.TempDir(),
		"XDG_CONFIG_HOME="+t.TempDir(),
		"AppData="+t.TempDir(),
	)
	if strings.Contains(r.stderr, "\033[") {
		t.Fatalf("expected no ANSI codes with NO_COLOR=1, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "No host specified") {
		t.Fatalf("expected error message without colors, got %q", r.stderr)
	}
}

func TestIntegration_SuccessUpload(t *testing.T) {
	if !sshAvailable(t) {
		t.Skip("ssh to localhost not available")
	}

	bin := buildBinary(t)
	pngPath := testPNGPath(t)
	putImageOnClipboard(t, pngPath)

	r := runBinary(t, bin, []string{"--host", "localhost", "--dir", "/tmp"})
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}

	if !strings.Contains(r.stderr, "Uploaded to") {
		t.Fatalf("expected 'Uploaded to' in stderr, got %q", r.stderr)
	}
	if !strings.Contains(r.stderr, "Path copied to clipboard") {
		t.Fatalf("expected 'Path copied to clipboard' in stderr, got %q", r.stderr)
	}

	remotePath := strings.TrimSpace(r.stdout)
	if !strings.HasPrefix(remotePath, "/tmp/scpclip_") || !strings.HasSuffix(remotePath, ".png") {
		t.Fatalf("unexpected remote path format: %q", remotePath)
	}

	if err := exec.Command("ssh", "localhost", "test", "-f", remotePath).Run(); err != nil {
		t.Fatalf("file not found on remote: %s", remotePath)
	}

	if runtime.GOOS == "linux" {
		out, err := exec.Command("ssh", "localhost", "stat", "-c", "%a", remotePath).Output()
		if err == nil {
			perms := strings.TrimSpace(string(out))
			if perms != "600" {
				t.Fatalf("expected permissions 600, got %s", perms)
			}
		}
	}

	clipText := readClipboardText(t)
	if clipText != remotePath {
		t.Fatalf("expected clipboard to contain %q, got %q", remotePath, clipText)
	}

	exec.Command("ssh", "localhost", "rm", "-f", remotePath).Run() //nolint:errcheck
}

func TestIntegration_PipeMode(t *testing.T) {
	if !sshAvailable(t) {
		t.Skip("ssh to localhost not available")
	}

	bin := buildBinary(t)
	pngPath := testPNGPath(t)
	putImageOnClipboard(t, pngPath)

	r := runBinary(t, bin, []string{"--host", "localhost", "--dir", "/tmp"})
	if r.exitCode != 0 {
		t.Fatalf("expected exit 0, got %d: stderr=%s", r.exitCode, r.stderr)
	}

	stdout := strings.TrimSpace(r.stdout)
	lines := strings.Split(stdout, "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly 1 line on stdout, got %d: %q", len(lines), stdout)
	}
	if strings.Contains(stdout, "✓") || strings.Contains(stdout, "\033[") {
		t.Fatalf("stdout should be raw path only, got %q", stdout)
	}

	exec.Command("ssh", "localhost", "rm", "-f", stdout).Run() //nolint:errcheck
}
