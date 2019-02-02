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

type VcsInfo struct {
	VcsName        string `json:"vcs_name" xml:"vcsName"`
	Path           string `json:"path" xml:"path"`
	RepositoryRoot string `json:"repository_root" xml:"repositoryRoot"`
	ShortHash      string `json:"short_hash" xml:"shortHash"`
	Hash           string `json:"hash" xml:"hash"`
	Revision       string `json:"revision" xml:"revision"`
	Branch         string `json:"branch" xml:"branch"`
	HasStaged      bool   `json:"has_staged" xml:"hasStaged"`
	HasModified    bool   `json:"has_modified" xml:"hasModified"`
	HasNew         bool   `json:"has_new" xml:"hasNew"`
}

type FormatOptions struct {
	HasStaged   string
	HasModified string
	HasNew      string
	Unknown     string
}

type VcsProbe interface {
	Name() string
	DefaultFormat() string
	IsAvailable() (bool, error)
	IsRepositoryRoot(path string) (bool, error)
	GatherInfo(path string) (VcsInfo, []error)
}

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

func FindProbeForPath(path string, probes []VcsProbe) (VcsProbe, error) {
	var goodProbe VcsProbe

	isAcceptable := func(path string) (bool, error) {
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
	return goodProbe, err
}

func InfoToJson(info VcsInfo) (string, error) {
	out, err := json.Marshal(info)
	return string(out[:]), err
}

func InfoToXml(info VcsInfo) (string, error) {
	out, err := xml.Marshal(info)
	return string(out[:]), err
}

func GetDefaultFormatOptions() FormatOptions {
	return FormatOptions{
		HasStaged:   "*",
		HasModified: "+",
		HasNew:      "?",
		Unknown:     "",
	}
}

func InfoToString(info VcsInfo, format string, options FormatOptions) (string, error) {
	var buf bytes.Buffer
	var eof rune = 0

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
