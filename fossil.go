package vcsinfo

import (
	"path/filepath"
	"strings"
)

// FossilProbe is a probe for extracting information out of a Fossil repository.
type FossilProbe struct{}

// Name returns the human-facing name of the probe.
func (probe FossilProbe) Name() string {
	return "fossil"
}

// DefaultFormat returns the default format string to use for Fossil
// repositories.
func (probe FossilProbe) DefaultFormat() string {
	return "%n[%b%m%u]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe FossilProbe) IsAvailable() (bool, error) {
	return commandExists("fossil"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of a Fossil repository.
func (probe FossilProbe) IsRepositoryRoot(path string) (bool, error) {
	return fileExists(filepath.Join(path, ".fslckout"))
}

func (probe FossilProbe) extractInfo(path string, info *VcsInfo) error {
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
}

func (probe FossilProbe) extractChanges(path string, info *VcsInfo) error {
	out, err := runCommand(path, "fossil", "changes")
	if err != nil {
		return err
	}

	if len(out) > 0 {
		info.HasModified = true
	}

	return nil
}

func (probe FossilProbe) extractExtras(path string, info *VcsInfo) error {
	out, err := runCommand(path, "fossil", "extras")
	if err != nil {
		return err
	}

	if len(out) > 0 {
		info.HasNew = true
	}

	return nil
}

// GatherInfo extracts and returns VCS information for the Fossil repository at
// the specified path.
func (probe FossilProbe) GatherInfo(path string) (VcsInfo, []error) {
	info := VcsInfo{
		VcsName: probe.Name(),
		Path:    path,
	}

	errors := waitGroup(
		func() error {
			return probe.extractInfo(path, &info)
		},

		func() error {
			return probe.extractChanges(path, &info)
		},

		func() error {
			return probe.extractExtras(path, &info)
		},
	)

	return info, errors
}
