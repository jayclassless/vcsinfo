package vcsinfo

import (
	"path/filepath"
	"strings"
)

type HgProbe struct{}

func (probe HgProbe) Name() string {
	return "hg"
}

func (probe HgProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

func (probe HgProbe) IsAvailable() (bool, error) {
	return commandExists("hg"), nil
}

func (probe HgProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".hg"))
}

func (probe HgProbe) GatherInfo(path string) (VcsInfo, []error) {
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
			out, err := runCommand(path, "hg", "status", "--modified", "--added", "--removed", "--unknown", "--removed", "--deleted")
			if err != nil {
				return err
			}

			for _, line := range out {
				if strings.HasPrefix(line, "?") {
					info.HasNew = true
				} else {
					info.HasModified = true
				}
			}

			return nil
		},

		func() error {
			out, err := runCommand(path, "hg", "identify", "--branch", "--num", "--id", "--debug")
			if err != nil {
				return err
			}

			parts := strings.Split(out[0], " ")

			info.Branch = parts[2]

			if strings.HasPrefix(parts[0], "0000000000000000000000000000000000000000") {
				return nil
			}

			info.Hash = parts[0]
			if strings.HasSuffix(info.Hash, "+") {
				info.Hash = info.Hash[0 : len(info.Hash)-1]
			}
			info.ShortHash = info.Hash[0:12]

			info.Revision = parts[1]
			if strings.HasSuffix(info.Revision, "+") {
				info.Revision = info.Revision[0 : len(info.Revision)-1]
			}

			return nil
		},
	)

	return info, errors
}
