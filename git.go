package vcsinfo

import (
	"path/filepath"
	"strings"
)

// GitProbe is a probe for extracting information out of a Git repository.
type GitProbe struct{}

// Name returns the human-facing name of the probe.
func (probe GitProbe) Name() string {
	return "git"
}

// DefaultFormat returns the default format string to use for Git repositories.
func (probe GitProbe) DefaultFormat() string {
	return "%n[%b%a%m%u%t]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe GitProbe) IsAvailable() (bool, error) {
	return commandExists("git"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of a Git repository.
func (probe GitProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".git"))
}

func (probe GitProbe) extractStatus(path string, info *VcsInfo) error {
	out, err := runCommand(path, "git", "status", "--porcelain")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode == 128 && strings.Contains(out[0], "must be run in a work tree") {
			return nil
		}
		return err
	}

	for _, line := range out {
		index, work := line[0:1], line[1:2]

		if index == "?" || work == "?" {
			info.HasNew = true
		} else {
			if index != " " {
				info.HasStaged = true
			}
			if work != " " {
				info.HasModified = true
			}
		}
	}

	return nil
}

func (probe GitProbe) extractBranch(path string, info *VcsInfo) error {
	out, err := runCommand(path, "git", "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return err
	}

	info.Branch = out[0]
	return nil
}

func (probe GitProbe) extractShortHash(path string, info *VcsInfo) error {
	out, err := runCommand(path, "git", "rev-parse", "--short", "HEAD")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode == 128 {
			// This generally means the repo doesn't have a commit yet.
			return nil
		}
		return err
	}

	info.ShortHash = out[0]
	return nil
}

func (probe GitProbe) extractHash(path string, info *VcsInfo) error {
	out, err := runCommand(path, "git", "rev-parse", "HEAD")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode == 128 {
			// This generally means the repo doesn't have a commit yet.
			return nil
		}
		return err
	}

	info.Hash = out[0]
	return nil
}

func (probe GitProbe) extractStashed(path string, info *VcsInfo) error {
	out, err := runCommand(path, "git", "stash", "list")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode == 1 && strings.Contains(out[0], "without a working tree") {
			// This generally means the repo doesn't have a commit yet.
			return nil
		}
		return err
	}

	info.HasStashed = len(out) > 0
	return nil
}

// GatherInfo extracts and returns VCS information for the Git repository at
// the specified path.
func (probe GitProbe) GatherInfo(path string) (VcsInfo, []error) {
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
			return probe.extractBranch(path, &info)
		},

		func() error {
			return probe.extractHash(path, &info)
		},

		func() error {
			return probe.extractShortHash(path, &info)
		},

		func() error {
			return probe.extractStashed(path, &info)
		},
	)

	return info, errors
}
