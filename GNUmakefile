TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: build

release:
	rm -fr bin
	mkdir -p bin
	GOARCH=amd64 GOOS=windows go build -o bin/terraform-provider-cf_windows_amd64.exe
	GOARCH=amd64 GOOS=linux go build -o bin/terraform-provider-cf_linux_amd64
	GOARCH=amd64 GOOS=darwin go build -o bin/terraform-provider-cf_darwin_amd64

build: check
	go install

test: check
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: check
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 180m

fmt:
	gofmt -w $(GOFMT_FILES)

check:
	@gometalinter.v2 --config .gometalinter.json --deadline 120s

vendor-status:
	@govendor status

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./aws"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: build test testacc fmt check vendor-status test-compile
