package vcsinfo_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	. "github.com/jayclassless/vcsinfo"
)


var _ = Describe("Bazaar", func() {
    probe := BzrProbe{}

    Describe("Name", func() {
        It("works", func() {
            Expect(probe.Name()).To(Equal("bzr"))
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
            run(dir, "bzr", "init-repo", ".")
            run(dir, "bzr", "init", "trunk")
            Expect(probe.IsRepositoryRoot(dir + "/trunk")).To(BeTrue())
        })
    })

    Describe("GatherInfo", func() {
        var dir, repoDir string

        BeforeEach(func() {
            repoDir = tmpdir()
            dir = repoDir + "/trunk"
            run(repoDir, "bzr", "init-repo", ".")
            run(repoDir, "bzr", "init", "trunk")
        })

        AfterEach(func() {
            rmdir(repoDir)
            dir = ""
            repoDir = ""
        })

        It("returns the basics", func() {
            info, err := probe.GatherInfo(dir)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "VcsName": Equal("bzr"),
                "Path": Equal(dir),
                "RepositoryRoot": Equal(dir),
                "Branch": Equal("trunk"),
            }))
        })

        It("returns the basics when deep in repo", func() {
            deep := mkdir(dir, "/some/deep/path")
            info, err := probe.GatherInfo(deep)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "VcsName": Equal("bzr"),
                "Path": Equal(deep),
                "RepositoryRoot": Equal(dir),
                "Branch": Equal("trunk"),
            }))
        })

        It("sees nothing when empty", func() {
            info, err := probe.GatherInfo(dir)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "HasNew": BeFalse(),
                "HasModified": BeFalse(),
                "HasStaged": BeFalse(),
                "Hash": Equal(""),
                "Revision": Equal(""),
                "Branch": Equal("trunk"),
            }))
        })

        It("sees new files", func() {
            writeFile(dir, "foo", "bar")
            info, err := probe.GatherInfo(dir)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "HasNew": BeTrue(),
                "HasModified": BeFalse(),
                "HasStaged": BeFalse(),
            }))
        })

        It("sees modified files", func() {
            writeFile(dir, "foo", "bar")
            run(dir, "bzr", "add", "foo")
            run(dir, "bzr", "commit", "-m", "blah")
            writeFile(dir, "foo", "baz")
            info, err := probe.GatherInfo(dir)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "HasNew": BeFalse(),
                "HasModified": BeTrue(),
                "HasStaged": BeFalse(),
                "Hash": Not(Equal("")),
            }))
        })

        It("sees deleted files", func() {
            writeFile(dir, "foo", "bar")
            run(dir, "bzr", "add", "foo")
            run(dir, "bzr", "commit", "-m", "blah")
            rm(dir, "foo")
            info, err := probe.GatherInfo(dir)
            Expect(err).To(BeEmpty())

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "HasNew": BeFalse(),
                "HasModified": BeTrue(),
                "HasStaged": BeFalse(),
                "Hash": Not(Equal("")),
            }))
        })

        It("sees branches", func() {
            run(repoDir, "bzr", "branch", "trunk", "mycoolbranch")
            info, _ := probe.GatherInfo(repoDir + "/mycoolbranch")

            Expect(info).To(MatchFields(IgnoreExtras, Fields{
                "Branch": Equal("mycoolbranch"),
            }))
        })
    })
})

