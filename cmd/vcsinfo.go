package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jayclassless/vcsinfo"
)

var (
	version = "dev"

	app = kingpin.New(
		"vcsinfo",
		"Retrieves and outputs basic information about the status of a VCS repository.",
	)

	helpFormat = app.Flag(
		"help-format",
		"Show help about defining output format strings",
	).Bool()
	helpEnvar = app.Flag(
		"help-envar",
		"Show help about environment variables that influence the execution of VCSInfo",
	).Bool()
	targetPath = app.Flag(
		"path",
		"The path to retrieve VCS information for.",
	).Short('p').String()
	format = app.Flag(
		"format",
		"The output format of the VCS information.",
	).Short('f').OverrideDefaultFromEnvar("VCSINFO_FORMAT").String()
	formatUntracked = app.Flag(
		"format-untracked",
		"The string to use for the untracked files indicator.",
	).Default("?").OverrideDefaultFromEnvar("VCSINFO_UNTRACKED").String()
	formatModified = app.Flag(
		"format-modified",
		"The string to use for the modified files indicator.",
	).Default("+").OverrideDefaultFromEnvar("VCSINFO_MODIFIED").String()
	formatStaged = app.Flag(
		"format-staged",
		"The string to use for the staged files indicator.",
	).Default("*").OverrideDefaultFromEnvar("VCSINFO_STAGED").String()
	formatStashed = app.Flag(
		"format-stashed",
		"The string to use for the stashed changes indicator.",
	).Default("@").OverrideDefaultFromEnvar("VCSINFO_STASHED").String()
	formatUnknown = app.Flag(
		"format-unknown",
		"The string to use for format codes where no value could be determined.",
	).Default("").OverrideDefaultFromEnvar("VCSINFO_UNKNOWN").String()
	json = app.Flag(
		"json",
		"Renders the output in a JSON object (overrides --format).",
	).Short('j').Bool()
	xml = app.Flag(
		"xml",
		"Renders the output in an XML document (overrides --format).",
	).Short('x').Bool()
	noisy = app.Flag(
		"noisy",
		"If hard failures are encountered, complain loudly instead of silently outputting nothing.",
	).Bool()

	probeFormats = make(map[string]*string)

	helpFormatText = `
The following codes can be used in format strings to embed VCS information
detected by VCSInfo in its output:

  %%n  VCS name
  %%h  Hash
  %%s  Short Hash
  %%r  Revision ID
  %%v  Short Hash, Revision ID, or Hash (whichever one that is found first is used)
  %%b  Branch
  %%u  Untracked files indicator
  %%a  Staged files indicator
  %%m  Modified files indicator
  %%t  Stashed changes indicator
  %%P  Repository root directory
  %%p  Relative path to Repository root directory (relative to the analyzed path)
  %%e  Base name of the repository root directory
  %%%%  Literal "%%"

If no format string is specified on the command line or via environment
variables, then the following strings will be used, depending on which VCS is
detected:

%s
`
	helpEnvarText = `
The following environment variables influence the execution and/or output of
VCSInfo:

  VCSINFO_FORMAT
    The format string to use to generate output (if not explicitly specified
    via the command line).

  VCSINFO_UNTRACKED
    The string to use for the untracked files indicator.

  VCSINFO_MODIFIED
    The string to use for the modified files indicator.

  VCSINFO_STAGED
    The string to use for the staged files indicator.

  VCSINFO_STASHED
    The string to used for the stashed changes indicator.

  VCSINFO_UNKNOWN
    The string to use for the %%h/%%s/%%r/%%s/%%v/%%b tokens if they could
    not be determined. Defaults to "".

%s
`
)

func determinePath() (string, error) {
	var path string
	var err error

	path = *targetPath
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	} else {
		if strings.HasPrefix(path, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			path = home + path[1:]
		}
	}

	path, err = filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", err
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	return path, nil
}

func produceOutput(info vcsinfo.VcsInfo, probe vcsinfo.VcsProbe) (string, error) {
	if *json {
		return vcsinfo.InfoToJSON(info)
	}
	if *xml {
		return vcsinfo.InfoToXML(info)
	}

	f := *probeFormats[probe.Name()]
	if f == "" {
		f = *format
		if f == "" {
			f = probe.DefaultFormat()
		}
	}

	options := vcsinfo.GetDefaultFormatOptions()
	options.HasNew = *formatUntracked
	options.HasModified = *formatModified
	options.HasStaged = *formatStaged
	options.HasStashed = *formatStashed
	options.Unknown = *formatUnknown

	return vcsinfo.InfoToString(info, f, options)
}

func failIfError(err error, message string) {
	if err != nil {
		if *noisy {
			app.FatalIfError(err, message)
		} else {
			os.Exit(0)
		}
	}
}

func makeDefaultFormatHelp(probes []vcsinfo.VcsProbe) string {
	m := make(map[string][]string)

	for _, probe := range probes {
		f := probe.DefaultFormat()
		_, exists := m[f]
		if exists {
			m[f] = append(m[f], probe.Name())
		} else {
			m[f] = []string{probe.Name()}
		}
	}

	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var out string
	for _, key := range keys {
		probes := m[key]
		sort.Strings(probes)
		out = fmt.Sprintf("%s  %s:\n    %s\n\n", out, strings.Join(probes, ", "), key)
	}

	return out
}

func makeEnvarHelp(probes []vcsinfo.VcsProbe) string {
	var out string

	for _, probe := range probes {
		out = fmt.Sprintf("%s  VCSINFO_FORMAT_%s\n    The format string to use to generate output for %s repositories.\n\n", out, strings.ToUpper(probe.Name()), probe.Name())
	}

	return out
}

func main() {
	allProbes, err := vcsinfo.GetAvailableProbes()
	failIfError(err, "Could not determine available VCS probes")

	for _, probe := range allProbes {
		probeFormats[probe.Name()] = app.Flag(
			fmt.Sprintf("format-%s", probe.Name()),
			fmt.Sprintf("The output format of the VCS information for %s repositories.", probe.Name()),
		).OverrideDefaultFromEnvar(
			fmt.Sprintf("VCSINFO_FORMAT_%s", strings.ToUpper(probe.Name())),
		).PlaceHolder("FORMAT").String()
	}

	app.Version(version)
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *helpFormat {
		fmt.Println(strings.TrimSpace(
			fmt.Sprintf(helpFormatText, makeDefaultFormatHelp(allProbes)),
		))
		os.Exit(0)
	}
	if *helpEnvar {
		fmt.Println(strings.TrimSpace(
			fmt.Sprintf(helpEnvarText, makeEnvarHelp(allProbes)),
		))
		os.Exit(0)
	}

	path, err := determinePath()
	failIfError(err, "Could not find path to analyze")

	probe, err := vcsinfo.FindProbeForPath(path, allProbes)
	failIfError(err, "Failure detecting VCS")

	if probe != nil {
		info, errs := probe.GatherInfo(path)
		if *noisy && len(errs) > 0 {
			for _, err := range errs {
				app.Errorf("%s", err)
			}
			app.Fatalf("Failure retrieving VCS information")
		}

		output, err := produceOutput(info, probe)
		failIfError(err, "Failure producing output")

		fmt.Println(output)
	}
}
