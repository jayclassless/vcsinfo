package vcsinfo

import (
	"path/filepath"
	"strings"
)

// HgProbe is a probe for extracting information out of a Mercurial repository.
type HgProbe struct{}

// Name returns the human-facing name of the probe.
func (probe HgProbe) Name() string {
	return "hg"
}

// DefaultFormat returns the default format string to use for Mercurial
// repositories.
func (probe HgProbe) DefaultFormat() string {
	return "%n[%b%m%u%t]"
}

// IsAvailable indicates whether or not this probe has the tools/environment
// necessary to operate.
func (probe HgProbe) IsAvailable() (bool, error) {
	return commandExists("hg"), nil
}

// IsRepositoryRoot identifies whether or not the specified path is the root
// of a Mercurial repository.
func (probe HgProbe) IsRepositoryRoot(path string) (bool, error) {
	return dirExists(filepath.Join(path, ".hg"))
}

func (probe HgProbe) extractStatus(path string, info *VcsInfo) error {
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
}

func (probe HgProbe) extractCommitInfo(path string, info *VcsInfo) error {
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
}

func (probe HgProbe) extractShelved(path string, info *VcsInfo) error {
	out, err := runCommand(path, "hg", "shelve", "--list")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode == 255 {
			// This generally means the shelve extension isn't enabled.
			return nil
		}
		return err
	}

	info.HasStashed = len(out) > 0
	return nil
}

// GatherInfo extracts and returns VCS information for the Mercurial repository
// at the specified path.
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
