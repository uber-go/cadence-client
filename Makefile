.PHONY: git-submodules test bins clean cover cover_ci

# default target
default: test

OS = $(shell uname -s)
ARCH = $(shell uname -m)

IMPORT_ROOT := go.uber.org/cadence/v2
THRIFT_GENDIR := .gen/go
THRIFTRW_SRC := idls/thrift/cadence.thrift idls/thrift/shadower.thrift
# one or more thriftrw-generated file(s), to create / depend on generated code
THRIFTRW_OUT := $(THRIFT_GENDIR)/cadence/idl.go
TEST_ARG ?= -v -race

# general build-product folder, cleaned as part of `make clean`
BUILD := .build
# general bins folder.  NOT cleaned via `make clean`
BINS := .bins

INTEG_TEST_ROOT := ./test
COVER_ROOT := $(BUILD)/coverage
UT_COVER_FILE := $(COVER_ROOT)/unit_test_cover.out
INTEG_STICKY_OFF_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_off_cover.out
INTEG_STICKY_ON_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_on_cover.out

# Automatically gather all srcs + a "sentinel" thriftrw output file (which forces generation).
ALL_SRC := $(THRIFTRW_OUT) $(shell \
	find . -name "*.go" | \
	grep -v \
	-e .gen/ \
	-e .build/ \
)

UT_DIRS := $(filter-out $(INTEG_TEST_ROOT)%, $(sort $(dir $(filter %_test.go,$(ALL_SRC)))))

# Files that needs to run lint.  excludes testify mocks and the thrift sentinel.
LINT_SRC := $(filter-out ./mock% ./tools.go $(THRIFTRW_OUT),$(ALL_SRC))

$(BINS)/thriftrw: go.mod
	go build -mod=readonly -o $@ go.uber.org/thriftrw

$(BINS)/thriftrw-plugin-yarpc: go.mod
	go build -mod=readonly -o $@ go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc

$(BINS)/golint: go.mod
	go build -mod=readonly -o $@ golang.org/x/lint/golint

$(BINS)/staticcheck: go.mod
	go build -mod=readonly -o $@ honnef.co/go/tools/cmd/staticcheck

$(BINS)/errcheck: go.mod
	go build -mod=readonly -o $@ github.com/kisielk/errcheck

$(BINS)/protoc-gen-gogofast: go.mod
	go build -mod=readonly -o $@ github.com/gogo/protobuf/protoc-gen-gogofast

$(BINS)/protoc-gen-yarpc-go: go.mod
	go build -mod=readonly -o $@ go.uber.org/yarpc/encoding/protobuf/protoc-gen-yarpc-go


$(THRIFTRW_OUT): $(THRIFTRW_SRC) $(BINS)/thriftrw $(BINS)/thriftrw-plugin-yarpc
	@echo 'thriftrw: $(THRIFTRW_SRC)'
	@mkdir -p $(dir $@)
	@# needs to be able to find the thriftrw-plugin-yarpc bin in PATH
	$(foreach source,$(THRIFTRW_SRC),\
		PATH="$(BINS)" \
		    $(BINS)/thriftrw \
		        --plugin=yarpc \
		        --pkg-prefix=$(IMPORT_ROOT)/$(THRIFT_GENDIR) \
		        --out=$(THRIFT_GENDIR) $(source);)

# https://www.grpc.io/docs/languages/go/quickstart/
# protoc-gen-gogofast (yarpc) are versioned via tools.go + go.mod (built above) and will be rebuilt as needed.
# changing PROTOC_VERSION will automatically download and use the specified version
PROTOC_VERSION = 3.14.0
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(subst Darwin,osx,$(OS))-$(ARCH).zip
# the zip contains an /include folder that we need to use to learn the well-known types
PROTOC_UNZIP_DIR = $(BINS)/protoc-$(PROTOC_VERSION)-zip
# use PROTOC_VERSION_BIN as a bin prerequisite, not "protoc", so the correct version will be used.
# otherwise this must be a .PHONY rule, or the buf bin / symlink could become out of date.
PROTOC_VERSION_BIN = protoc-$(PROTOC_VERSION)
$(BINS)/$(PROTOC_VERSION_BIN): | $(BINS)
	@echo "downloading protoc $(PROTOC_VERSION)"
	@# recover from partial success
	@rm -rf $(BINS)/protoc.zip $(PROTOC_UNZIP_DIR)
	@# download, unzip, copy to a normal location
	@curl -sSL $(PROTOC_URL) -o $(BINS)/protoc.zip
	@unzip -q $(BINS)/protoc.zip -d $(PROTOC_UNZIP_DIR)
	@cp $(PROTOC_UNZIP_DIR)/bin/protoc $@

PROTO_ROOT := idls/proto
PROTO_OUT := .gen/proto
PROTO_FILES = $(shell find ./$(PROTO_ROOT) -name "*.proto")

protoc: $(PROTO_FILES) $(BINS)/$(PROTOC_VERSION_BIN) $(BINS)/protoc-gen-gogofast $(BINS)/protoc-gen-yarpc-go
	@mkdir -p $(PROTO_OUT)
	@echo "protoc..."
	@$(BINS)/$(PROTOC_VERSION_BIN) \
		--plugin $(BINS)/protoc-gen-gogofast \
		--plugin $(BINS)/protoc-gen-yarpc-go \
		-I=$(PROTO_ROOT) \
		-I=$(PROTOC_UNZIP_DIR)/include \
		--gogofast_out=Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,paths=source_relative:$(PROTO_OUT) \
		--yarpc-go_out=$(PROTO_OUT) \
		$(PROTO_FILES);
	@rm -r $(PROTO_OUT)/api
	@mv $(PROTO_OUT)/uber/cadence/api $(PROTO_OUT)
	@rm -r $(PROTO_OUT)/uber

git-submodules:
	git submodule update --init --recursive

yarpc-install: $(BINS)/thriftrw $(BINS)/thriftrw-plugin-yarpc

thriftc: git-submodules yarpc-install $(THRIFTRW_OUT) copyright

clean_thrift:
	rm -rf .gen

# `make copyright` or depend on "copyright" to force-run licensegen,
# or depend on $(BUILD)/copyright to let it run as needed.
copyright $(BUILD)/copyright: $(ALL_SRC)
	@mkdir -p $(BUILD)
	go run ./internal/cmd/tools/copyright/licensegen.go --verifyOnly
	@touch $(BUILD)/copyright

$(BUILD)/dummy:
	go build -o $@ internal/cmd/dummy/dummy.go

bins: thriftc $(ALL_SRC) $(BUILD)/copyright lint $(BUILD)/dummy

unit_test: $(BUILD)/dummy
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(UT_COVER_FILE)
	@for dir in $(UT_DIRS); do \
		mkdir -p $(COVER_ROOT)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(COVER_ROOT)/"$$dir"/cover.out || exit 1; \
		cat $(COVER_ROOT)/"$$dir"/cover.out | grep -v "mode: atomic" >> $(UT_COVER_FILE); \
	done;

integ_test_sticky_off: $(BUILD)/dummy
	@mkdir -p $(COVER_ROOT)
	STICKY_OFF=true go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_OFF_COVER_FILE) -coverpkg=./...

integ_test_sticky_on: $(BUILD)/dummy
	@mkdir -p $(COVER_ROOT)
	STICKY_OFF=false go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_ON_COVER_FILE) -coverpkg=./...

test: thriftc unit_test integ_test_sticky_off integ_test_sticky_on

$(COVER_ROOT)/cover.out: $(UT_COVER_FILE) $(INTEG_STICKY_OFF_COVER_FILE) $(INTEG_STICKY_ON_COVER_FILE)
	@echo "mode: atomic" > $(COVER_ROOT)/cover.out
	cat $(UT_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_OFF_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_ON_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out

cover: $(COVER_ROOT)/cover.out
	go tool cover -html=$(COVER_ROOT)/cover.out;

cover_ci: $(COVER_ROOT)/cover.out
	goveralls -coverprofile=$(COVER_ROOT)/cover.out -service=buildkite || echo -e "\x1b[31mCoveralls failed\x1b[m";

# golint fails to report many lint failures if it is only given a single file
# to work on at a time, and it can't handle multiple packages at once, *and*
# we can't exclude files from its checks, so for best results we need to give
# it a whitelist of every file in every package that we want linted, per package.
#
# so lint + this golint func works like:
# - iterate over all lintable dirs (outputs "./folder/")
# - find .go files in a dir (via wildcard, so not recursively)
# - filter to only files in LINT_SRC
# - if it's not empty, run golint against the list
define lint_if_present
test -n "$1" && $(BINS)/golint -set_exit_status $1
endef

lint: $(BINS)/golint $(ALL_SRC)
	$(foreach pkg,\
		$(sort $(dir $(LINT_SRC))), \
		$(call lint_if_present,$(filter $(wildcard $(pkg)*.go),$(LINT_SRC))) || ERR=1; \
	) test -z "$$ERR" || exit 1
	@OUTPUT=`gofmt -l $(ALL_SRC) 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "Run 'make fmt'. gofmt must be run on the following files:"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi

staticcheck: $(BINS)/staticcheck $(ALL_SRC)
	$(BINS)/staticcheck ./...

errcheck: $(BINS)/errcheck $(ALL_SRC)
	$(BINS)/errcheck ./...

fmt:
	@gofmt -w $(ALL_SRC)

clean:
	rm -Rf $(BUILD)
	rm -Rf .gen

# broken up into multiple += so I can interleave comments.
# this all becomes a single line of output.
# you must not use single-quotes within the string in this var.
JQ_DEPS_AGE = jq '
# only deal with things with updates
JQ_DEPS_AGE += select(.Update)
# allow additional filtering, e.g. DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)'
JQ_DEPS_AGE += $(DEPS_FILTER)
# add "days between current version and latest version"
JQ_DEPS_AGE += | . + {Age:(((.Update.Time | fromdate) - (.Time | fromdate))/60/60/24 | floor)}
# add "days between latest version and now"
JQ_DEPS_AGE += | . + {Available:((now - (.Update.Time | fromdate))/60/60/24 | floor)}
# 123 days: library 	old_version -> new_version
JQ_DEPS_AGE += | ([.Age, .Available] | max | tostring) + " days: " + .Path + "  \t" + .Version + " -> " + .Update.Version
JQ_DEPS_AGE += '
# remove surrounding quotes from output
JQ_DEPS_AGE += --raw-output

# exclude `"Indirect": true` dependencies.  direct ones have no "Indirect" key at all.
JQ_DEPS_ONLY_DIRECT = | select(has("Indirect") | not)

deps: ## Check for dependency updates, for things that are directly imported
	@make --no-print-directory DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)' deps-all

deps-all: ## Check for all dependency updates
	@go list -u -m -json all \
		| $(JQ_DEPS_AGE) \
		| sort -n
