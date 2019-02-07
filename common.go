package vcsinfo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// VcsInfo contains the results of a VcsProbe's examination of a repository.
type VcsInfo struct {
	// The name of the VCS found.
	VcsName string `json:"vcs_name" xml:"vcsName"`

	// The path that was examined.
	Path string `json:"path" xml:"path"`

	// The root directory of the repository that was examined.
	RepositoryRoot string `json:"repository_root" xml:"repositoryRoot"`

	// The "short" version of the Hash of the current changeset, if the VCS has
	// such a concept.
	ShortHash string `json:"short_hash" xml:"shortHash"`

	// The hash of the current changeset.
	Hash string `json:"hash" xml:"hash"`

	// The revision ID of the current changeset.
	Revision string `json:"revision" xml:"revision"`

	// The current branch.
	Branch string `json:"branch" xml:"branch"`

	// Indicates whether or not there are files staged for commit.
	HasStaged bool `json:"has_staged" xml:"hasStaged"`

	// Indicates whether or not there are added/modified/deleted files.
	HasModified bool `json:"has_modified" xml:"hasModified"`

	// Indicates whether or not there are untracked files.
	HasNew bool `json:"has_new" xml:"hasNew"`
}

// FormatOptions contains the options that govern how format strings are
// produced.
type FormatOptions struct {
	// The string displayed for the staged file indicator.
	HasStaged string

	// The string displayed for the modified file indicator.
	HasModified string

	// The string displayed for the untracked file indicator.
	HasNew string

	// The string displayed for hash/rev/branch tokens when the information
	// they represent could not be found.
	Unknown string
}

// VcsProbe represents a probe that is capable of examining the current state
// of a VCS repository.
type VcsProbe interface {
	// Name returns the human-facing name of the probe.
	Name() string

	// DefaultFormat returns the default format string to use for the
	// repositories.
	DefaultFormat() string

	// IsAvailable indicates whether or not this probe has the
	// tools/environment necessary to operate.
	IsAvailable() (bool, error)

	// IsRepositoryRoot identifies whether or not the specified path is the root
	// of a repository this probe can handle.
	IsRepositoryRoot(path string) (bool, error)

	// GatherInfo extracts and returns VCS information for the repository at
	// the specified path.
	GatherInfo(path string) (VcsInfo, []error)
}

// GetAvailableProbes returns all probes in the VCSInfo package that can used
// in the current environment.
func GetAvailableProbes() ([]VcsProbe, error) {
	allProbes := []VcsProbe{
		GitProbe{},
		HgProbe{},
		SvnProbe{},
		BzrProbe{},
		FossilProbe{},
		DarcsProbe{},
		CvsProbe{},
	}

	availableProbes := []VcsProbe{}

	for _, probe := range allProbes {
		available, err := probe.IsAvailable()
		if err != nil {
			return nil, err
		}
		if available {
			availableProbes = append(availableProbes, probe)
		}
	}

	return availableProbes, nil
}

// FindProbeForPath identifies which of the specified VcsProbes is appropriate
// to use to examine the specified path.
func FindProbeForPath(path string, probes []VcsProbe) (VcsProbe, error) {
	var goodProbe VcsProbe

	isAcceptable := func(path string) (bool, error) {
		skipExists, err := fileExists(filepath.Join(path, ".novcsinfo"))
		if err != nil {
			return false, err
		} else if skipExists {
			return false, fmt.Errorf("novcsinfo")
		}

		for _, probe := range probes {
			exists, err := probe.IsRepositoryRoot(path)
			if err != nil {
				return false, err
			}
			if exists {
				goodProbe = probe
				return true, nil
			}
		}
		return false, nil
	}

	_, err := findAcceptablePath(path, isAcceptable)
	if err != nil && err.Error() == "novcsinfo" {
		return nil, nil
	}

	return goodProbe, err
}

// InfoToJSON renders the VcsInfo as a JSON object.
func InfoToJSON(info VcsInfo) (string, error) {
	out, err := json.Marshal(info)
	return string(out[:]), err
}

// InfoToXML renders the VcsInfo as an XML document.
func InfoToXML(info VcsInfo) (string, error) {
	out, err := xml.Marshal(info)
	return string(out[:]), err
}

// GetDefaultFormatOptions returns a FormatOptions initialized with the default
// configuration.
func GetDefaultFormatOptions() FormatOptions {
	return FormatOptions{
		HasStaged:   "*",
		HasModified: "+",
		HasNew:      "?",
		Unknown:     "",
	}
}

// InfoToString renders the VcsInfo according the specified format string and
// options.
func InfoToString(info VcsInfo, format string, options FormatOptions) (string, error) {
	var buf bytes.Buffer
	var eof rune

	reader := bufio.NewReader(strings.NewReader(format))

	sou := func(str string) string {
		if str == "" {
			return options.Unknown
		}
		return str
	}

	for {
		char, _, _ := reader.ReadRune()
		if char == eof {
			break
		}

		if char != '%' {
			buf.WriteString(string(char))
			continue
		}

		char, _, _ = reader.ReadRune()
		switch char {
		case 'n':
			buf.WriteString(info.VcsName)

		case 'h':
			buf.WriteString(sou(info.Hash))

		case 's':
			buf.WriteString(sou(info.ShortHash))

		case 'r':
			buf.WriteString(sou(info.Revision))

		case 'v':
			out := options.Unknown
			if info.ShortHash != "" {
				buf.WriteString(info.ShortHash)
			} else if info.Revision != "" {
				buf.WriteString(info.Revision)
			} else if info.Hash != "" {
				buf.WriteString(info.Hash)
			}
			buf.WriteString(out)

		case 'b':
			buf.WriteString(sou(info.Branch))

		case 'u':
			if info.HasNew {
				buf.WriteString(options.HasNew)
			}

		case 'a':
			if info.HasStaged {
				buf.WriteString(options.HasStaged)
			}

		case 'm':
			if info.HasModified {
				buf.WriteString(options.HasModified)
			}

		case 'P':
			buf.WriteString(info.RepositoryRoot)

		case 'p':
			relPath, err := filepath.Rel(info.RepositoryRoot, info.Path)
			if err == nil {
				buf.WriteString(relPath)
			}

		case 'e':
			buf.WriteString(path.Base(info.RepositoryRoot))

		case '%':
			buf.WriteString("%")

		default:
			return "", fmt.Errorf("Unexpected formatting code \"%%%s\"", string(char))
		}
	}

	return buf.String(), nil
}
