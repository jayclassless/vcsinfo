package vcsinfo_test

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)


func TestProbes(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "vcsinfo Suite")
}

func mkdir(pathParts ...string) string {
    path := filepath.Join(pathParts...)
    os.MkdirAll(path, os.ModePerm)
    return path
}

func tmpdir() string {
    dir, _ := ioutil.TempDir("", "vcsinfo")
    return dir
}

func writeFile(dir string, file string, content string) {
    ioutil.WriteFile(filepath.Join(dir, file), []byte(content), 0666)
}

func rm(dir string, file string) {
    os.Remove(filepath.Join(dir, file))
}

func rmdir(dir string) {
    os.RemoveAll(dir)
}

func run(dir string, command ...string) {
    cmd := exec.Command(command[0], command[1:]...)
    cmd.Dir = dir
    err := cmd.Run()
    if err != nil {
        Fail(fmt.Sprintf("Failed to execute %s %+v", command, err))
    }
}

