package vcsinfo

import (
	pth "path"
	"path/filepath"
	"strings"
)

type DarcsProbe struct{}

func (probe DarcsProbe) Name() string {
	return "darcs"
}

func (probe DarcsProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

func (probe DarcsProbe) IsAvailable() (bool, error) {
	return commandExists("darcs"), nil
}

func (probe DarcsProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, "_darcs"))
}

func (probe DarcsProbe) GatherInfo(path string) (VcsInfo, []error) {
	info := VcsInfo{
		VcsName: probe.Name(),
		Path:    path,
	}

	root, err := findAcceptablePath(path, probe.IsRepositoryRoot)
	if err != nil {
		return info, []error{err}
	}
	info.RepositoryRoot = root
	info.Branch = pth.Base(root)

	errors := waitGroup(
		func() error {
			out, err := runCommand(path, "darcs", "whatsnew", "--look-for-adds", "--summary")
			if err != nil {
				if len(out) > 0 {
					if out[0] == "No changes!" {
						return nil
					}
				}
				return err
			}

			for _, line := range out {
				flag := line[0:1]

				if flag == "a" {
					info.HasNew = true
				} else {
					info.HasModified = true
				}
			}

			return nil
		},

		func() error {
			out, err := runCommand(path, "darcs", "log", "--last", "1")
			if err != nil {
				return err
			}

			for _, line := range out {
				if strings.HasPrefix(line, "patch ") {
					info.Hash = line[6:]
				}
			}

			return nil
		},
	)

	return info, errors
}
