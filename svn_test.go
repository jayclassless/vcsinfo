package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jayclassless/vcsinfo"
)

var _ = Describe("Subversion", func() {
	probe := SvnProbe{}

	Describe("Name", func() {
		It("works", func() {
			Expect(probe.Name()).To(Equal("svn"))
		})
	})

	Describe("DefaultFormat", func() {
		It("works", func() {
			Expect(probe.DefaultFormat()).To(Not(Equal("")))
		})
	})

	Describe("IsAvailable", func() {
		It("works", func() {
			Expect(probe.IsAvailable()).To(BeTrue())
		})
	})

	Describe("IsRepositoryRoot", func() {
		var dir, repoDir, repoUrl string

		BeforeEach(func() {
			dir = tmpdir()
			repoDir = tmpdir()
			repoUrl = "file://" + repoDir + "/TestRepo"
			run(repoDir, "svnadmin", "create", "TestRepo")
			run(dir, "svn", "mkdir", "-m", "dirs", repoUrl+"/trunk", repoUrl+"/branches", repoUrl+"/tags")
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
			rmdir(repoDir)
			repoDir = ""
		})

		It("returns false in dir with no repo", func() {
			Expect(probe.IsRepositoryRoot(dir)).To(BeFalse())
		})

		It("returns true in dir with new repo", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			Expect(probe.IsRepositoryRoot(dir)).To(BeTrue())
		})
	})

	Describe("GatherInfo", func() {
		var dir, repoDir, repoUrl string

		BeforeEach(func() {
			dir = tmpdir()
			repoDir = tmpdir()
			repoUrl = "file://" + repoDir + "/TestRepo"
			run(repoDir, "svnadmin", "create", "TestRepo")
			run(dir, "svn", "mkdir", "-m", "dirs", repoUrl+"/trunk", repoUrl+"/branches", repoUrl+"/tags")
			run(dir, "svn", "copy", repoUrl+"/trunk", repoUrl+"/branches/mybranch", "-m", "Creating a new branch")
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
			rmdir(repoDir)
			repoDir = ""
		})

		It("returns the basics", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("svn"),
				"Path":           Equal(dir),
				"RepositoryRoot": Equal(dir),
				"Revision":       Equal("1"),
				"Branch":         Equal("trunk"),
			}))
		})

		It("returns the basics when deep in repo", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			deep := mkdir(dir, "/some/deep/path")
			info, err := probe.GatherInfo(deep)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("svn"),
				"Path":           Equal(deep),
				"RepositoryRoot": Equal(dir),
				"Revision":       Equal(""),
				"Branch":         Equal(""),
			}))
		})

		It("sees nothing when empty", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
				"Hash":        Equal(""),
				"ShortHash":   Equal(""),
				"Revision":    Equal("1"),
				"Branch":      Equal("trunk"),
			}))
		})

		It("sees new files", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			writeFile(dir, "foo", "bar")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeTrue(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
			}))
		})

		It("sees modified files", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			writeFile(dir, "foo", "bar")
			run(dir, "svn", "add", "foo")
			run(dir, "svn", "commit", "-m", "blah")
			writeFile(dir, "foo", "baz")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"Revision":    Not(Equal("")),
			}))
		})

		It("sees deleted files", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			writeFile(dir, "foo", "bar")
			run(dir, "svn", "add", "foo")
			run(dir, "svn", "commit", "-m", "blah")
			rm(dir, "foo")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"Revision":    Not(Equal("")),
			}))
		})

		It("sees branches", func() {
			run(dir, "svn", "checkout", repoUrl+"/branches/mybranch", ".")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"Revision": Equal("2"),
				"Branch":   Equal("mybranch"),
			}))
		})

		It("doesnt crash when in VCS special dir", func() {
			run(dir, "svn", "checkout", repoUrl+"/trunk", ".")
			_, err := probe.GatherInfo(dir + "/.svn")
			Expect(err).To(BeEmpty())
		})
	})
})
