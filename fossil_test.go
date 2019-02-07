package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"path/filepath"
	"time"

	. "github.com/jayclassless/vcsinfo"
)

var _ = Describe("Fossil", func() {
	probe := FossilProbe{}

	Describe("Name", func() {
		It("works", func() {
			Expect(probe.Name()).To(Equal("fossil"))
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
		var dir, repoDir, repo string

		BeforeEach(func() {
			dir = tmpdir()
			repoDir = tmpdir()
			run(repoDir, "fossil", "init", "foorepo")
			repo = repoDir + "/foorepo"
		})

		AfterEach(func() {
			rmdir(dir)
			rmdir(repoDir)
			dir = ""
			repoDir = ""
			repo = ""
		})

		It("returns false in dir with no repo", func() {
			Expect(probe.IsRepositoryRoot(dir)).To(BeFalse())
		})

		It("returns true in dir with new repo", func() {
			run(dir, "fossil", "open", repo)
			Expect(probe.IsRepositoryRoot(dir)).To(BeTrue())
		})
	})

	Describe("GatherInfo", func() {
		var dir, repoDir, repo string

		BeforeEach(func() {
			dir = tmpdir()
			repoDir = tmpdir()
			run(repoDir, "fossil", "init", "foorepo")
			repo = repoDir + "/foorepo"
			run(dir, "fossil", "open", repo)
		})

		AfterEach(func() {
			rmdir(dir)
			rmdir(repoDir)
			dir = ""
			repoDir = ""
			repo = ""
		})

		It("returns the basics", func() {
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			fullDir, _ := filepath.EvalSymlinks(dir)
			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("fossil"),
				"Path":           Equal(dir),
				"RepositoryRoot": Equal(fullDir + "/"),
				"Branch":         Equal("trunk"),
			}))
		})

		It("returns the basics when deep in repo", func() {
			deep := mkdir(dir, "/some/deep/path")
			info, err := probe.GatherInfo(deep)
			Expect(err).To(BeEmpty())

			fullDir, _ := filepath.EvalSymlinks(dir)
			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("fossil"),
				"Path":           Equal(deep),
				"RepositoryRoot": Equal(fullDir + "/"),
				"Branch":         Equal("trunk"),
			}))
		})

		It("sees nothing when empty", func() {
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
				"Hash":        Not(Equal("")),
				"Branch":      Equal("trunk"),
			}))
		})

		It("sees new files", func() {
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
			writeFile(dir, "foo", "bar")
			run(dir, "fossil", "add", "foo")
			run(dir, "fossil", "commit", "-m", "blah")

			// Darcs "file has been updated" detection is poor. If we update the file too quickly, it doesn't see it as a change
			time.Sleep(1001 * time.Millisecond)
			writeFile(dir, "foo", "baz")

			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"Hash":        Not(Equal("")),
			}))
		})

		It("sees deleted files", func() {
			writeFile(dir, "foo", "bar")
			run(dir, "fossil", "add", "foo")
			run(dir, "fossil", "commit", "-m", "blah")
			rm(dir, "foo")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"Hash":        Not(Equal("")),
			}))
		})

		It("sees branches", func() {
			writeFile(dir, "bar", "baz")
			run(dir, "fossil", "add", "bar")
			run(dir, "fossil", "commit", "-m", "blah", "--branch", "foobranch")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"Branch": Equal("foobranch"),
			}))
		})
	})
})
