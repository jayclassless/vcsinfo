GOBIN = ${shell go env GOPATH}/bin


init::
	@go get github.com/onsi/ginkgo/ginkgo
	@go get golang.org/x/lint/golint
	@go get github.com/fzipp/gocyclo
	@go get github.com/mattn/goveralls

test::
	@${GOBIN}/ginkgo -p -cover -coverprofile=coverage.out

coverage::
	@go tool cover -html=coverage.out

lint::
	@${GOBIN}/golint ./...
	@${GOBIN}/gocyclo -over 10 .

fmt::
	@gofmt -s -w .

test-publish::
	@goreleaser release --snapshot --rm-dist

clean::
	@rm -rf dist coverage.out

