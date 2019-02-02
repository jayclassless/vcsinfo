package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jayclassless/vcsinfo"
)

var _ = Describe("CVS", func() {
	probe := CvsProbe{}

	Describe("Name", func() {
		It("works", func() {
			Expect(probe.Name()).To(Equal("cvs"))
		})
	})

	Describe("IsAvailable", func() {
		It("works", func() {
			Expect(probe.IsAvailable()).To(BeTrue())
		})
	})

	Describe("IsRepositoryRoot", func() {
		var dir, repoDir string

		cvs := func(targetDir string, command ...string) {
			cmd := append([]string{"cvs", "-d", repoDir}, command...)
			run(targetDir, cmd...)
		}

		BeforeEach(func() {
			dir = tmpdir()

			repoDir = tmpdir()
			cvs(repoDir, "init")

			dummy := mkdir(dir, "dummy")
			cvs(dir, "import", "-m", "Initial import", "dummy", "mycompany", "init")
			rmdir(dummy)
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
			cvs(dir, "checkout", "dummy", ".")
			Expect(probe.IsRepositoryRoot(dir)).To(BeTrue())
		})
	})

	Describe("GatherInfo", func() {
		var dir, repoDir string

		cvs := func(targetDir string, command ...string) {
			cmd := append([]string{"cvs", "-d", repoDir}, command...)
			run(targetDir, cmd...)
		}

		BeforeEach(func() {
			dir = tmpdir()

			repoDir = tmpdir()
			cvs(repoDir, "init")

			dummy := mkdir(dir, "dummy")
			cvs(dir, "import", "-m", "Initial import", "dummy", "mycompany", "init")
			rmdir(dummy)
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
			rmdir(repoDir)
			repoDir = ""
		})

		It("returns the basics", func() {
			cvs(dir, "checkout", "dummy", ".")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("cvs"),
				"Path":           Equal(dir),
				"RepositoryRoot": Equal(dir),
				"Branch":         Equal(""),
			}))
		})

		It("returns the basics when deep in repo", func() {
			cvs(dir, "checkout", "dummy", ".")

			mkdir(dir, "/some")
			cvs(dir, "add", "some")

			deep := mkdir(dir, "/some/deep/path")
			info, err := probe.GatherInfo(deep)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("cvs"),
				"Path":           Equal(deep),
				"RepositoryRoot": Equal(dir),
				"Branch":         Equal(""),
			}))
		})

		It("sees nothing when empty", func() {
			cvs(dir, "checkout", "dummy", ".")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
			}))
		})

		It("sees new files", func() {
			cvs(dir, "checkout", "dummy", ".")
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
			cvs(dir, "checkout", "dummy", ".")
			writeFile(dir, "foo", "bar")
			cvs(dir, "add", "foo")
			cvs(dir, "commit", "-m", "blah")
			writeFile(dir, "foo", "baz")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"Hash":        Equal(""),
				"ShortHash":   Equal(""),
				"Revision":    Equal(""),
			}))
		})

		It("sees deleted files", func() {
			cvs(dir, "checkout", "dummy", ".")
			writeFile(dir, "foo", "bar")
			cvs(dir, "add", "foo")
			cvs(dir, "commit", "-m", "blah")
			rm(dir, "foo")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
			}))
		})
	})
})
