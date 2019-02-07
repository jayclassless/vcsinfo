package vcsinfo

import (
	"path/filepath"
	"strings"
)

// BzrProbe is a probe for extracting information out of an Bazaar repository.
type BzrProbe struct{}

// Name returns the human-facing name of the probe.
func (probe BzrProbe) Name() string {
	return "bzr"
}

// DefaultFormat returns the default format string to use for Bazaar
// repositories.
func (probe BzrProbe) DefaultFormat() string {
	return "%n[%b%m%u%t]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe BzrProbe) IsAvailable() (bool, error) {
	return commandExists("bzr"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of an Bazaar repository.
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

func (probe BzrProbe) extractShelved(path string, info *VcsInfo) error {
	_, err := runCommand(path, "bzr", "shelve", "--list")
	if err != nil {
		exitCode := getExitCode(err)
		info.HasStashed = exitCode > 0
	}
	return nil
}

// GatherInfo extracts and returns VCS information for the Bazaar repository at
// the specified path.
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

		func() error {
			return probe.extractShelved(path, &info)
		},
	)

	return info, errors
}
