package vcsinfo

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

func dirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}

	if fileInfo.IsDir() {
		return true, nil
	}

	return false, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}

	return true, nil
}

func commandExists(command string) bool {
	exists, err := exec.LookPath(command)
	return exists != "" && err == nil
}

func getExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return -1
}

func runCommand(workingDir string, command ...string) ([]string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = workingDir
	out, err := cmd.CombinedOutput()

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, err
}

func waitGroup(routines ...func() error) []error {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(routines))

	errors := []error{}

	for idx := range routines {
		routine := routines[idx]
		go func() {
			defer waitGroup.Done()
			err := routine()
			if err != nil {
				errors = append(errors, err)
			}
		}()
	}

	waitGroup.Wait()

	return errors
}

func findAcceptablePath(path string, isAcceptable func(string) (bool, error)) (string, error) {
	for {
		acceptable, err := isAcceptable(path)
		if err != nil {
			return "", err
		}

		if acceptable {
			return path, nil
		}

		if path == "/" {
			break
		}
		path = filepath.Join(path, "..")
	}

	return "", nil
}
