package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jayclassless/vcsinfo"
)

var _ = Describe("Mercurial", func() {
	probe := HgProbe{}

	Describe("Name", func() {
		It("works", func() {
			Expect(probe.Name()).To(Equal("hg"))
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
		var dir string

		BeforeEach(func() {
			dir = tmpdir()
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
		})

		It("returns false in dir with no repo", func() {
			Expect(probe.IsRepositoryRoot(dir)).To(BeFalse())
		})

		It("returns true in dir with new repo", func() {
			run(dir, "hg", "init")
			Expect(probe.IsRepositoryRoot(dir)).To(BeTrue())
		})
	})

	Describe("GatherInfo", func() {
		var dir string

		BeforeEach(func() {
			dir = tmpdir()
			run(dir, "hg", "init")
		})

		AfterEach(func() {
			rmdir(dir)
			dir = ""
		})

		It("returns the basics", func() {
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("hg"),
				"Path":           Equal(dir),
				"RepositoryRoot": Equal(dir),
				"Branch":         Equal("default"),
			}))
		})

		It("returns the basics when deep in repo", func() {
			deep := mkdir(dir, "/some/deep/path")
			info, err := probe.GatherInfo(deep)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"VcsName":        Equal("hg"),
				"Path":           Equal(deep),
				"RepositoryRoot": Equal(dir),
				"Branch":         Equal("default"),
			}))
		})

		It("sees nothing when empty", func() {
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
				"HasStashed":  BeFalse(),
				"Hash":        Equal(""),
				"ShortHash":   Equal(""),
				"Revision":    Equal(""),
				"Branch":      Equal("default"),
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
				"HasStashed":  BeFalse(),
			}))
		})

		It("sees modified files", func() {
			writeFile(dir, "foo", "bar")
			run(dir, "hg", "add", "foo")
			run(dir, "hg", "commit", "-m", "blah")
			writeFile(dir, "foo", "baz")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"HasStashed":  BeFalse(),
				"Hash":        Not(Equal("")),
				"ShortHash":   Not(Equal("")),
			}))
		})

		It("sees deleted files", func() {
			writeFile(dir, "foo", "bar")
			run(dir, "hg", "add", "foo")
			run(dir, "hg", "commit", "-m", "blah")
			rm(dir, "foo")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeTrue(),
				"HasStaged":   BeFalse(),
				"HasStashed":  BeFalse(),
				"Hash":        Not(Equal("")),
				"ShortHash":   Not(Equal("")),
			}))
		})

		It("sees stashed changes", func() {
			writeFile(dir, "foo", "bar")
			run(dir, "hg", "add", "foo")
			run(dir, "hg", "shelve")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"HasNew":      BeFalse(),
				"HasModified": BeFalse(),
				"HasStaged":   BeFalse(),
				"HasStashed":  BeTrue(),
			}))
		})

		It("sees branches", func() {
			run(dir, "hg", "branch", "foo")
			info, err := probe.GatherInfo(dir)
			Expect(err).To(BeEmpty())

			Expect(info).To(MatchFields(IgnoreExtras, Fields{
				"Branch": Equal("foo"),
			}))
		})

		It("doesnt crash when in VCS special dir", func() {
			_, err := probe.GatherInfo(dir + "/.hg")
			Expect(err).To(BeEmpty())
		})
	})
})
