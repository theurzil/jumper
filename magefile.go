//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binary = "jumper"
	prefix = "/usr/local/bin"
	rcFile = ".bashrc"
	shDest = ".local/share/jumper/jumper.sh"
)

func version() (string, error) {
	b, err := os.ReadFile("VERSION")
	if err != nil {
		return "", err
	}
	return "v" + strings.TrimSpace(string(b)), nil
}

// Build compiles the jumper binary.
func Build() error {
	v, err := version()
	if err != nil {
		return err
	}
	return sh.RunV("go", "build", "-ldflags", "-X main.version="+v, "-o", binary, ".")
}

// Test runs the test suite with race detection.
func Test() error {
	return sh.RunV("go", "test", "-race", "-v", "./...")
}

// Install builds and installs the binary, shell script, and rc hook.
func Install() error {
	mg.Deps(Build)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dest := filepath.Join(prefix, binary)
	if writable(prefix) {
		if err := sh.Copy(dest, binary); err != nil {
			return err
		}
		if err := os.Chmod(dest, 0o755); err != nil {
			return err
		}
	} else {
		fmt.Println("No write access to", prefix, "using sudo...")
		if err := sh.RunV("sudo", "install", "-Dm755", binary, dest); err != nil {
			return err
		}
	}

	scriptDest := filepath.Join(home, shDest)
	if err := os.MkdirAll(filepath.Dir(scriptDest), 0o755); err != nil {
		return err
	}
	if err := sh.Copy(scriptDest, "jumper.sh"); err != nil {
		return err
	}

	rcPath := filepath.Join(home, rcFile)
	sourceLine := "source " + scriptDest
	if !containsLine(rcPath, sourceLine) {
		f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString("\n" + sourceLine + "\n"); err != nil {
			return err
		}
		fmt.Println("Added source line to", rcPath)
	}

	fmt.Println("Installed. Run: source", rcPath, " (or restart your shell)")
	return nil
}

// Uninstall removes the installed binary and shell script.
func Uninstall() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dest := filepath.Join(prefix, binary)
	if _, err := os.Stat(dest); err == nil {
		if writable(dest) {
			if err := os.Remove(dest); err != nil {
				return err
			}
		} else if err := sh.RunV("sudo", "rm", "-f", dest); err != nil {
			return err
		}
	}

	installDir := filepath.Join(home, ".local", "share", "jumper")
	if err := os.RemoveAll(installDir); err != nil {
		return err
	}

	rcPath := filepath.Join(home, rcFile)
	fmt.Println("Removed binary and", installDir)
	fmt.Println("Manually remove the 'source", filepath.Join(home, shDest), "' line from", rcPath)
	return nil
}

// Clean removes build artifacts.
func Clean() error {
	return sh.Rm(binary)
}

func writable(path string) bool {
	target := path
	if _, err := os.Stat(path); err != nil {
		target = filepath.Dir(path)
	}
	return unix_W_OK(target)
}

func unix_W_OK(path string) bool {
	cmd := exec.Command("test", "-w", path)
	return cmd.Run() == nil
}

func containsLine(path, line string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), line)
}
