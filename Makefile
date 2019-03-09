GOBIN = ${shell go env GOPATH}/bin

init::
	@go mod download
	@go install github.com/onsi/ginkgo/ginkgo
	@go install github.com/mattn/goveralls

test::
	@${GOBIN}/ginkgo -p -cover -coverprofile=coverage.out

ci-gha::
	${MAKE} init
	bzr whoami "Fake Tester <fake@example.com>"
	git config --global user.email "fake@example.com"
	git config --global user.name "Fake Tester"
	echo "[extensions]\nshelve=" > ~/.hgrc
	${MAKE} test

coverage::
	@go tool cover -html=coverage.out

lint::
	@golangci-lint run

fmt::
	@gofmt -s -w *.go cmd

test-publish::
	@goreleaser release --snapshot --rm-dist

clean::
	@-chmod -R u+w .vendor
	@rm -rf dist coverage.out .vendor

