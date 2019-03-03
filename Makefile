GOBIN = ${shell go env GOPATH}/bin

init::
	@go mod download
	@go install github.com/onsi/ginkgo/ginkgo
	@go install golang.org/x/lint/golint
	@go install github.com/fzipp/gocyclo
	@go install github.com/mattn/goveralls

test::
	@${GOBIN}/ginkgo -p -cover -coverprofile=coverage.out

coverage::
	@go tool cover -html=coverage.out

lint::
	@${GOBIN}/golint -set_exit_status . cmd
	@${GOBIN}/gocyclo -over 10 *.go cmd

fmt::
	@gofmt -s -w *.go cmd

test-publish::
	@goreleaser release --snapshot --rm-dist

clean::
	@-chmod -R u+w .vendor
	@rm -rf dist coverage.out .vendor

