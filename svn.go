package vcsinfo

import (
	"path/filepath"
	"strings"
)

// SvnProbe is a probe for extracting information out of an SVN repository.
type SvnProbe struct{}

// Name returns the human-facing name of the probe.
func (probe SvnProbe) Name() string {
	return "svn"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe SvnProbe) IsAvailable() (bool, error) {
	return commandExists("svn"), nil
}

// DefaultFormat returns the default format string to use for SVN repositories.
func (probe SvnProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of an SVN repository.
func (probe SvnProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".svn"))
}

func (probe SvnProbe) extractStatus(path string, info *VcsInfo) error {
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
}

func (probe SvnProbe) extractInfo(path string, info *VcsInfo) error {
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
			parts := strings.SplitN(line, "^", 2)
			if parts[1] == "/trunk" {
				info.Branch = "trunk"
			} else if strings.HasPrefix(parts[1], "/branches/") {
				info.Branch = parts[1][10:]
			}

		} else if strings.HasPrefix(line, "Last Changed Rev: ") {
			parts := strings.SplitN(line, ": ", 2)
			info.Revision = parts[1]
		}
	}

	return nil
}

// GatherInfo extracts and returns VCS information for the SVN repository at
// the specified path.
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
			return probe.extractStatus(path, &info)
		},

		func() error {
			return probe.extractInfo(path, &info)
		},
	)

	return info, errors
}
