.PHONY: test bins clean cover cover_ci

# default target
default: test

IMPORT_ROOT := go.uber.org/cadence
THRIFT_GENDIR := .gen/go
THRIFTRW_SRC := idl/github.com/uber/cadence/cadence.thrift
# one or more thriftrw-generated file(s), to create / depend on generated code
THRIFTRW_OUT := $(THRIFT_GENDIR)/cadence/idl.go
TEST_ARG ?= -coverprofile=$(BUILD)/cover.out -race

# general build-product folder, cleaned as part of `make clean`
BUILD := .build

$(THRIFTRW_OUT): $(THRIFTRW_SRC) $(BUILD)/thriftrw $(BUILD)/thriftrw-plugin-yarpc
	@echo 'thriftrw: $(THRIFTRW_SRC)'
	@mkdir -p $(dir $@)
	@# needs to be able to find the thriftrw-plugin-yarpc bin in PATH
	@PATH="$(BUILD)" \
		$(BUILD)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(IMPORT_ROOT)/$(THRIFT_GENDIR) \
		--out=$(THRIFT_GENDIR) \
		$(THRIFTRW_SRC)

# Automatically gather all srcs + a "sentinel" thriftrw output file (which forces generation).
ALL_SRC := $(THRIFTRW_OUT) $(shell \
	find . -name "*.go" | \
	grep -v \
	-e vendor/ \
	-e .gen/ \
	-e .build/ \
)

# Files that needs to run lint, exclude testify mock from lint
LINT_SRC := $(filter-out ./mock%,$(ALL_SRC))

vendor: vendor/glide.updated

vendor/glide.updated: glide.lock
	glide install
	touch vendor/glide.updated

$(BUILD)/thriftrw:
	./versioned_go_build.sh go.uber.org/thriftrw v1.11.0 $@

$(BUILD)/thriftrw-plugin-yarpc:
	./versioned_go_build.sh go.uber.org/yarpc v1.29.1 encoding/thrift/thriftrw-plugin-yarpc $@

$(BUILD)/golint:
	./versioned_go_build.sh golang.org/x/lint 470b6b0bb3005eda157f0275e2e4895055396a81 golint $@

clean_thrift:
	rm -rf .gen

# `make copyright` or depend on "copyright" to force-run licensegen,
# or depend on $(BUILD)/copyright to let it run as needed.
copyright $(BUILD)/copyright: $(ALL_SRC)
	go run ./internal/cmd/tools/copyright/licensegen.go --verifyOnly
	touch $(BUILD)/copyright

$(BUILD)/dummy: vendor/glide.updated $(ALL_SRC)
	go build -i -o $@ internal/cmd/dummy/dummy.go

test $(BUILD)/cover.out: $(BUILD)/copyright $(BUILD)/dummy $(ALL_SRC)
	go test ./... $(TEST_ARG)

bins: $(ALL_SRC) $(BUILD)/copyright lint $(BUILD)/dummy

cover: $(BUILD)/cover.out
	go tool cover -html=$(BUILD)/cover.out;

cover_ci: $(BUILD)/cover.out
	goveralls -coverprofile=$(BUILD)/cover.out -service=travis-ci || echo -e "\x1b[31mCoveralls failed\x1b[m";

# golint fails to report many lint failures if it is only given a single file
# to work on at a time.  and we can't exclude files from its checks, so for
# best results we need to give it a whitelist of every file in every package
# that we want linted.
#
# so lint + this golint func works like:
# - iterate over all dirs (outputs "./folder/")
# - find .go files in a dir (via wildcard, so not recursively)
# - filter to only files in LINT_SRC
# - if it's not empty, run golint against the list
define lint_if_present
test -n "$1" && $(BUILD)/golint -set_exit_status $1
endef

lint: $(BUILD)/golint $(ALL_SRC)
	@$(foreach pkg,\
		$(sort $(dir $(LINT_SRC))), \
		$(call lint_if_present,$(filter $(wildcard $(pkg)*.go),$(LINT_SRC))) || ERR=1; \
	) test -z "$$ERR" || exit 1
	@OUTPUT=`gofmt -l $(ALL_SRC) 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "Run 'make fmt'. gofmt must be run on the following files:"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi

fmt:
	@gofmt -w $(ALL_SRC)

clean:
	rm -Rf $(BUILD)
	rm -Rf .gen
