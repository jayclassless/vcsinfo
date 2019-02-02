package vcsinfo

import (
	"path/filepath"
	"strings"
)

type SvnProbe struct{}

func (probe SvnProbe) Name() string {
	return "svn"
}

func (probe SvnProbe) IsAvailable() (bool, error) {
	return commandExists("svn"), nil
}

func (probe SvnProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

func (probe SvnProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".svn"))
}

func (probe SvnProbe) GatherInfo(path string) (VcsInfo, []error) {
	info := VcsInfo{
		VcsName: probe.Name(),
		Path:    path,
	}

	root, err := findAcceptablePath(path, probe.IsRepositoryRoot)
	if err != nil {
		return info, []error{err}
	}
	info.RepositoryRoot = root

	errors := waitGroup(
		func() error {
			out, err := runCommand(path, "svn", "status")
			if err != nil {
				return err
			}

			for _, line := range out {
				item, props := line[0:1], line[1:2]

				if item == "?" {
					info.HasNew = true
				} else if item != " " || props != " " {
					info.HasModified = true
				}
			}

			return nil
		},

		func() error {
			out, err := runCommand(path, "svn", "info")
			if err != nil {
				if len(out) > 0 {
					// We're likely in a new directory that hasn't been added yet
					if strings.HasPrefix(out[len(out)-1], "svn: E200009") {
						return nil
					}
				}
				return err
			}

			for _, line := range out {
				if strings.HasPrefix(line, "Relative URL: ^") {
					parts := strings.Split(line, "^")
					if len(parts) != 2 {
						continue
					}

					if parts[1] == "/trunk" {
						info.Branch = "trunk"
					} else if strings.HasPrefix(parts[1], "/branches/") {
						info.Branch = parts[1][10:]
					}

				} else if strings.HasPrefix(line, "Last Changed Rev: ") {
					parts := strings.Split(line, ": ")
					if len(parts) != 2 {
						continue
					}

					info.Revision = parts[1]
				}
			}

			return nil
		},
	)

	return info, errors
}
