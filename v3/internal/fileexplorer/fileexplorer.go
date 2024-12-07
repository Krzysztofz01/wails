package fileexplorer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func Open(path string, selectFile bool) error {
	if pathInfo, err := os.Stat(path); err != nil {
		return fmt.Errorf("failed to access the specified path stat: %w", err)
	} else {
		selectFile = selectFile && !pathInfo.IsDir()
	}

	var (
		explorerBinArgs explorerBinArgs
		ignoreExitCode  bool = false
	)

	switch runtime.GOOS {
	case "windows":
		explorerBinArgs = windowsExplorerBinArgs

		// NOTE: Disabling the exit code check on Windows system. Workaround for explorer.exe
		// exit code handling (https://github.com/microsoft/WSL/issues/6565)
		ignoreExitCode = true
	case "darwin":
		explorerBinArgs = darwinExplorerBinArgs
	case "linux":
		explorerBinArgs = linuxExplorerBinArgs
	default:
		return errors.New("unsupported platform")
	}

	explorerBin, explorerArgs, err := explorerBinArgs(path, selectFile)
	if err != nil {
		return fmt.Errorf("failed to determine the file explorer binary: %w", err)
	}

	cmd := exec.Command(explorerBin, explorerArgs...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start the file explorer process: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok && ignoreExitCode {
			return nil
		}

		return fmt.Errorf("file explorer process failed: %w", err)
	}

	return nil
}

type explorerBinArgs = func(path string, selectFile bool) (string, []string, error)

var windowsExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	args := []string{}
	if selectFile {
		args = append(args, fmt.Sprintf("/select,\"%s\"", path))
	} else {
		args = append(args, path)
	}

	return "explorer", args, nil
}

var darwinExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	args := []string{}
	if selectFile {
		args = append(args, "-R")
	}

	args = append(args, path)

	return "open", args, nil
}

var linuxExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	var explorerBinArgs explorerBinArgs

	desktopEnvironment := strings.ToUpper(strings.TrimSpace(os.Getenv("XDG_CURRENT_DESKTOP")))
	switch desktopEnvironment {
	case "CINNAMON", "X-CINNAMON":
		explorerBinArgs = linuxCinnamonExplorerBinArgs
	case "GNOME", "GNOME-FLASHBACK", "GNOME-FLASHBACK:GNOME":
		explorerBinArgs = linuxGnomeExplorerBinArgs
	case "KDE":
		explorerBinArgs = linuxKdeExplorerBinArgs
	case "LXQT":
		explorerBinArgs = linuxLxqtExplorerBinArgs
	case "MATE":
		explorerBinArgs = linuxMateExplorerBinArgs
	case "XFCE":
		explorerBinArgs = linuxXfceExplorerBinArgs
	default:
		explorerBinArgs = linuxFallbackExplorerBinArgs
	}

	return explorerBinArgs(path, selectFile)
}

var linuxGnomeExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "nautilus", []string{path}, nil
}

var linuxKdeExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "dolphin", []string{path}, nil
}

var linuxXfceExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "thunar", []string{path}, nil
}

var linuxMateExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "caja", []string{path}, nil
}

var linuxLxqtExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "pcmanfm-qt", []string{path}, nil
}

var linuxCinnamonExplorerBinArgs explorerBinArgs = func(path string, selectFile bool) (string, []string, error) {
	if !selectFile {
		path = filepath.Dir(path)
	}

	return "nemo", []string{path}, nil
}

var linuxFallbackExplorerBinArgs explorerBinArgs = func(path string, _ bool) (string, []string, error) {
	// NOTE: The linux fallback explorer opening is not supporting file selection
	path = filepath.Dir(path)

	return "xdg-open", []string{path}, nil
}
