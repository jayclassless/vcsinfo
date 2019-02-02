GOBIN = ${shell go env GOPATH}/bin


init::
	@go get github.com/onsi/ginkgo/ginkgo
	@go get golang.org/x/lint/golint

test::
	@${GOBIN}/ginkgo -p -cover -coverprofile=coverage.out

coverage::
	@go tool cover -html=coverage.out

lint::
	@${GOBIN}/golint ./...

fmt::
	@gofmt -s -w .

