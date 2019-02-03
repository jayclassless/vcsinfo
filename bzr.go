package vcsinfo

import (
	"path/filepath"
	"strings"
)

type BzrProbe struct{}

func (probe BzrProbe) Name() string {
	return "bzr"
}

func (probe BzrProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

func (probe BzrProbe) IsAvailable() (bool, error) {
	return commandExists("bzr"), nil
}

func (probe BzrProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".bzr/branch"))
}

func (probe BzrProbe) extractStatus(path string, info *VcsInfo) error {
	out, err := runCommand(path, "bzr", "status")
	if err != nil {
		return err
	}

	for _, line := range out {
		if strings.HasPrefix(line, "added") ||
			strings.HasPrefix(line, "removed") ||
			strings.HasPrefix(line, "renamed") ||
			strings.HasPrefix(line, "kind changed") ||
			strings.HasPrefix(line, "modified") {
			info.HasModified = true
		} else if strings.HasPrefix(line, "unknown") {
			info.HasNew = true
		}
	}

	return nil
}

func (probe BzrProbe) extractCommitInfo(path string, info *VcsInfo) error {
	out, err := runCommand(path, "bzr", "version-info")
	if err != nil {
		return err
	}

	for _, line := range out {
		parts := strings.Split(line, ": ")
		if len(parts) != 2 {
			continue
		}

		if parts[0] == "revision-id" {
			info.Hash = parts[1]

		} else if parts[0] == "revno" && parts[1] != "0" {
			info.Revision = parts[1]

		} else if parts[0] == "branch-nick" {
			info.Branch = parts[1]

		}
	}

	return nil
}

func (probe BzrProbe) GatherInfo(path string) (VcsInfo, []error) {
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
			return probe.extractStatus(path, &info)
		},

		func() error {
			return probe.extractCommitInfo(path, &info)
		},
	)

	return info, errors
}
