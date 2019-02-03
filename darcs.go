package vcsinfo

import (
	pth "path"
	"path/filepath"
	"strings"
)

// DarcsProbe is a probe for extracting information out of a DARCS repository.
type DarcsProbe struct{}

// Name returns the human-facing name of the probe.
func (probe DarcsProbe) Name() string {
	return "darcs"
}

// DefaultFormat returns the default format string to use for DARCS
// repositories.
func (probe DarcsProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe DarcsProbe) IsAvailable() (bool, error) {
	return commandExists("darcs"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of a DARCS repository.
func (probe DarcsProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, "_darcs"))
}

func (probe DarcsProbe) extractStatus(path string, info *VcsInfo) error {
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
}

func (probe DarcsProbe) extractHash(path string, info *VcsInfo) error {
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
}

// GatherInfo extracts and returns VCS information for the DARCS repository at
// the specified path.
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
			return probe.extractStatus(path, &info)
		},

		func() error {
			return probe.extractHash(path, &info)
		},
	)

	return info, errors
}
