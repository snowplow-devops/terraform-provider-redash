.PHONY: all format lint tidy test clean

# -----------------------------------------------------------------------------
#  CONSTANTS
# -----------------------------------------------------------------------------

version = `cat VERSION`

src_dir := terraform-provider-redash

build_dir = build

coverage_dir  = $(build_dir)/coverage
coverage_out  = $(coverage_dir)/coverage.out
coverage_html = $(coverage_dir)/coverage.html

output_dir    = $(build_dir)/output

linux_dir     = $(output_dir)/linux
darwin_dir    = $(output_dir)/darwin
windows_dir   = $(output_dir)/windows

bin_name      = terraform-provider-redash_v$(version)
bin_linux     = $(linux_dir)/$(bin_name)
bin_darwin    = $(darwin_dir)/$(bin_name)
bin_windows   = $(windows_dir)/$(bin_name)

# -----------------------------------------------------------------------------
#  BUILDING
# -----------------------------------------------------------------------------

all:
	GO111MODULE=on go get -u github.com/mitchellh/gox
	GO111MODULE=on gox -osarch=linux/amd64 -output=$(bin_linux) ./$(src_dir)
	GO111MODULE=on gox -osarch=darwin/amd64 -output=$(bin_darwin) ./$(src_dir)
	GO111MODULE=on gox -osarch=windows/amd64 -output=$(bin_windows) ./$(src_dir)

# -----------------------------------------------------------------------------
#  FORMATTING
# -----------------------------------------------------------------------------

format:
	GO111MODULE=on go fmt ./$(src_dir)
	GO111MODULE=on gofmt -s -w ./$(src_dir)

lint:
	GO111MODULE=on go get -u golang.org/x/lint/golint
	GO111MODULE=on golint ./$(src_dir)

tidy:
	GO111MODULE=on go mod tidy

# -----------------------------------------------------------------------------
#  TESTING
# -----------------------------------------------------------------------------

test:
	mkdir -p $(coverage_dir)
	GO111MODULE=on go get -u golang.org/x/tools/cmd/cover/...
	GO111MODULE=on go test ./$(src_dir) -tags test -v -covermode=count -coverprofile=$(coverage_out)
	GO111MODULE=on go tool cover -html=$(coverage_out) -o $(coverage_html)

# -----------------------------------------------------------------------------
#  CLEANUP
# -----------------------------------------------------------------------------

clean:
	rm -rf $(build_dir)
	