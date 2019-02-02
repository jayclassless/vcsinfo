package vcsinfo

import (
    "path/filepath"
    "strings"
)


type FossilProbe struct {}


func (probe FossilProbe) Name() string {
    return "fossil"
}


func (probe FossilProbe) DefaultFormat() string {
    return "%n[%b%m%u]"
}


func (probe FossilProbe) IsAvailable() (bool, error) {
    return commandExists("fossil"), nil
}


func (probe FossilProbe) IsRepositoryRoot(path string) (bool, error) {
    return fileExists(filepath.Join(path, ".fslckout"))
}


func (probe FossilProbe) GatherInfo(path string) (VcsInfo, []error) {
    info := VcsInfo{
        VcsName: probe.Name(),
        Path: path,
    }

    errors := waitGroup(
        func() error {
            out, err := runCommand(path, "fossil", "info")
            if err != nil {
                return err
            }

            for _, line := range out {
                parts := strings.SplitN(line, ":", 2)
                if len(parts) != 2 {
                    continue
                }
                field := parts[0]
                value := strings.TrimSpace(parts[1])

                if field == "local-root" {
                    info.RepositoryRoot = value

                } else if field == "checkout" {
                    subparts := strings.SplitN(value, " ", 2)
                    info.Hash = subparts[0]

                } else if field == "tags" {
                    subparts := strings.SplitN(value, ", ", 2)
                    if len(subparts) == 0 {
                        continue
                    }
                    info.Branch = subparts[0]
                }
            }

            return nil
        },

        func() error {
            out, err := runCommand(path, "fossil", "changes")
            if err != nil {
                return err
            }

            if len(out) > 0 {
                info.HasModified = true
            }

            return nil
        },

        func() error {
            out, err := runCommand(path, "fossil", "extras")
            if err != nil {
                return err
            }

            if len(out) > 0 {
                info.HasNew = true
            }

            return nil
        },
   )

    return info, errors
}

