package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/jayclassless/vcsinfo"
)

var _ = Describe("Public API", func() {
	Describe("GetAvailableProbes", func() {
		It("returns all probes", func() {
			probes, err := GetAvailableProbes()
			Expect(err).To(BeNil())
			Expect(probes).To(HaveLen(7))
		})
	})

	Describe("FindProbeForPath", func() {
		var dir string
		var allProbes []VcsProbe

		BeforeEach(func() {
			dir = tmpdir()
			allProbes, _ = GetAvailableProbes()
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
			allProbes = nil
		})

		It("finds nothing", func() {
			probe, _ := FindProbeForPath(dir, allProbes)
			Expect(probe).To(BeNil())
		})

		It("finds a git repo when asked", func() {
			run(dir, "git", "init")
			probe, _ := FindProbeForPath(dir, allProbes)
			Expect(probe.Name()).To(Equal("git"))

			mkdir(dir, "/deeper/dir")
			probe, _ = FindProbeForPath(dir+"/deeper/dir", allProbes)
			Expect(probe.Name()).To(Equal("git"))
		})

		It("doesn't find a git repo when not asked", func() {
			run(dir, "git", "init")
			probe, _ := FindProbeForPath(dir, []VcsProbe{HgProbe{}, SvnProbe{}})
			Expect(probe).To(BeNil())
		})

		It("finds a hg repo when asked", func() {
			run(dir, "hg", "init")
			probe, _ := FindProbeForPath(dir, allProbes)
			Expect(probe.Name()).To(Equal("hg"))

			mkdir(dir, "/deeper/dir")
			probe, _ = FindProbeForPath(dir+"/deeper/dir", allProbes)
			Expect(probe.Name()).To(Equal("hg"))
		})

		It("honors .novcsinfo", func() {
			run(dir, "git", "init")
			mkdir(dir, "/deeper/dir")
			writeFile(dir, "/deeper/.novcsinfo", "")

			probe, _ := FindProbeForPath(dir+"/deeper/dir", allProbes)
			Expect(probe).To(BeNil())

			probe, _ = FindProbeForPath(dir+"/deeper", allProbes)
			Expect(probe).To(BeNil())

			probe, _ = FindProbeForPath(dir, allProbes)
			Expect(probe.Name()).To(Equal("git"))
		})
	})

	Describe("InfoToJSON", func() {
		It("renders to string", func() {
			info := VcsInfo{
				VcsName:        "fake",
				Path:           "/foo/bar",
				RepositoryRoot: "/foo",
				Hash:           "abc123",
				HasModified:    true,
			}
			actual, err := InfoToJSON(info)
			Expect(err).To(BeNil())
			Expect(actual).To(Equal(`{"vcs_name":"fake","path":"/foo/bar","repository_root":"/foo","short_hash":"","hash":"abc123","revision":"","branch":"","has_staged":false,"has_modified":true,"has_new":false}`))
		})
	})

	Describe("InfoToXML", func() {
		It("renders to string", func() {
			info := VcsInfo{
				VcsName:        "fake",
				Path:           "/foo/bar",
				RepositoryRoot: "/foo",
				Hash:           "abc123",
				HasModified:    true,
			}
			actual, err := InfoToXML(info)
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("<VcsInfo><vcsName>fake</vcsName><path>/foo/bar</path><repositoryRoot>/foo</repositoryRoot><shortHash></shortHash><hash>abc123</hash><revision></revision><branch></branch><hasStaged>false</hasStaged><hasModified>true</hasModified><hasNew>false</hasNew></VcsInfo>"))
		})
	})

	Describe("InfoToString", func() {
		It("renders to string", func() {
			info := VcsInfo{
				VcsName:        "fake",
				Path:           "/foo/bar",
				RepositoryRoot: "/foo",
				Hash:           "abc123",
				ShortHash:      "xyz",
				Revision:       "42",
				Branch:         "master",
				HasModified:    true,
				HasNew:         true,
				HasStaged:      true,
			}
			actual, err := InfoToString(info, "%%|%n|%h|%s|%r|%v|%b|%u|%a|%m|%P|%p|%e", GetDefaultFormatOptions())
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("%|fake|abc123|xyz|42|xyz|master|?|*|+|/foo|bar|foo"))
		})

		It("handles the %v fallbacks", func() {
			info := VcsInfo{
				Hash:      "abc123",
				ShortHash: "xyz",
				Revision:  "42",
			}
			actual, err := InfoToString(info, "%v", GetDefaultFormatOptions())
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("xyz"))

			info = VcsInfo{
				Hash:     "abc123",
				Revision: "42",
			}
			actual, err = InfoToString(info, "%v", GetDefaultFormatOptions())
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("42"))

			info = VcsInfo{
				Hash: "abc123",
			}
			actual, err = InfoToString(info, "%v", GetDefaultFormatOptions())
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("abc123"))

			info = VcsInfo{}
			actual, err = InfoToString(info, "%v", GetDefaultFormatOptions())
			Expect(err).To(BeNil())
			Expect(actual).To(Equal(""))
		})

		It("handles changed options", func() {
			info := VcsInfo{
				HasNew:      true,
				HasModified: true,
				HasStaged:   true,
			}
			options := GetDefaultFormatOptions()
			options.Unknown = "dunno"
			options.HasNew = "@"
			options.HasModified = "#"
			options.HasStaged = "$"

			actual, err := InfoToString(info, "%h|%s|%r|%v|%b|%u|%a|%m", options)
			Expect(err).To(BeNil())
			Expect(actual).To(Equal("dunno|dunno|dunno|dunno|dunno|@|$|#"))
		})

		It("fails on unrecognized codes", func() {
			info := VcsInfo{}
			actual, err := InfoToString(info, "%Q", GetDefaultFormatOptions())
			Expect(actual).To(Equal(""))
			Expect(err).To(MatchError(`Unexpected formatting code "%Q"`))
		})
	})
})
