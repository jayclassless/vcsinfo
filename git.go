package vcsinfo

import (
    "path/filepath"
)


type GitProbe struct {}


func (probe GitProbe) Name() string {
    return "git"
}


func (probe GitProbe) DefaultFormat() string {
    return "%n[%b%a%m%u]"
}


func (probe GitProbe) IsAvailable() (bool, error) {
    return commandExists("git"), nil
}


func (probe GitProbe) IsRepositoryRoot(path string) (bool, error) {
    return dirExists(filepath.Join(path, ".git"))
}


func (probe GitProbe) GatherInfo(path string) (VcsInfo, []error) {
    info := VcsInfo{
        VcsName: probe.Name(),
        Path: path,
    }

    root, err := findAcceptablePath(path, probe.IsRepositoryRoot)
    if err != nil {
        return info, []error{err}
    }
    info.RepositoryRoot = root

    errors := waitGroup(
        func() error {
            out, err := runCommand(path, "git", "status", "--porcelain")
            if err != nil {
                return err
            }

            for _, line := range out {
                index, work := line[0:1], line[1:2]

                if index == "?" || work == "?" {
                    info.HasNew = true
                } else {
                    if index != " " {
                        info.HasStaged = true
                    }
                    if work != " " {
                        info.HasModified = true
                    }
                }
            }

            return nil
        },

        func() error {
            out, err := runCommand(path, "git", "rev-parse", "--short", "HEAD")
            if err != nil {
                exitCode := getExitCode(err)
                if exitCode == 128 {
                    // This generally means the repo doesn't have a commit yet.
                    return nil
                }
                return err
            }

            info.ShortHash = out[0]
            return nil
        },

        func() error {
            out, err := runCommand(path, "git", "rev-parse", "HEAD")
            if err != nil {
                exitCode := getExitCode(err)
                if exitCode == 128 {
                    // This generally means the repo doesn't have a commit yet.
                    return nil
                }
                return err
            }

            info.Hash = out[0]
            return nil
        },

        func() error {
            out, err := runCommand(path, "git", "symbolic-ref", "--short", "HEAD")
            if err != nil {
                return err
            }

            info.Branch = out[0]
            return nil
        },
   )

    return info, errors
}

