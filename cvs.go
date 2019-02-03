package vcsinfo

import (
	"path/filepath"
	"strings"
)

// CvsProbe is a probe for extracting information out of a CVS  repository.
type CvsProbe struct{}

// Name returns the human-facing name of the probe.
func (probe CvsProbe) Name() string {
	return "cvs"
}

// DefaultFormat returns the default format string to use for CVS repositories.
func (probe CvsProbe) DefaultFormat() string {
	return "%n[%e%m%u]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe CvsProbe) IsAvailable() (bool, error) {
	return commandExists("cvs"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of a CVS repository.
func (probe CvsProbe) IsRepositoryRoot(path string) (bool, error) {
	exists, err := dirExists(filepath.Join(path, "CVS"))
	if !exists || err != nil {
		return false, err
	}

	parentExists, parentErr := dirExists(filepath.Join(path, "..", "CVS"))
	if parentExists || parentErr != nil {
		return false, parentErr
	}

	return true, nil
}

func (probe CvsProbe) extractStatus(path string, info *VcsInfo) error {
	out, err := runCommand(path, "cvs", "status")
	if err != nil {
		if len(out) > 0 {
			// We're likely in a new directory that hasn't been added yet
			if strings.HasPrefix(out[0], "cvs status: No CVSROOT specified!") {
				return nil
			}
		}
		return err
	}

	for _, line := range out {
		if strings.HasSuffix(line, "Locally Added") ||
			strings.HasSuffix(line, "Locally Modified") ||
			strings.HasSuffix(line, "Locally Removed") ||
			strings.HasSuffix(line, "Needs Checkout") {
			info.HasModified = true
		}
	}

	return nil
}

func (probe CvsProbe) extractNew(path string, info *VcsInfo) error {
	out, err := runCommand(path, "cvs", "-qn", "update")
	if err != nil {
		if len(out) > 0 {
			// We're likely in a new directory that hasn't been added yet
			if strings.HasPrefix(out[0], "cvs update: No CVSROOT specified!") {
				return nil
			}
		}
		return err
	}

	for _, line := range out {
		if strings.HasPrefix(line, "?") {
			info.HasNew = true
			return nil
		}
	}

	return nil
}

// GatherInfo extracts and returns VCS information for the CVS repository at
// the specified path.
func (probe CvsProbe) GatherInfo(path string) (VcsInfo, []error) {
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
			return probe.extractNew(path, &info)
		},
	)

	return info, errors
}
