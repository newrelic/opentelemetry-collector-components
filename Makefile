include ./Makefile.Common

RUN_CONFIG?=local/config.yaml
CMD?=
OTEL_VERSION=main

BUILD_INFO_IMPORT_PATH=github.com/cristianciutea/opentelemetry-components/internal/otelcomponents/internal/version
VERSION=$(shell git describe --always --match "v[0-9]*" HEAD)
BUILD_INFO=-ldflags "-X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION)"

COMP_REL_PATH=internal/components/components.go
MOD_NAME=github.com/cristianciutea/opentelemetry-components

GROUP ?= all
FOR_GROUP_TARGET=for-$(GROUP)-target

FIND_MOD_ARGS=-type f -name "go.mod"
TO_MOD_DIR=dirname {} \; | sort | grep -E '^./'
EX_COMPONENTS=-not -path "./receiver/*" -not -path "./processor/*" -not -path "./exporter/*" -not -path "./extension/*"
EX_INTERNAL=-not -path "./internal/*"

# NONROOT_MODS includes ./* dirs (excludes . dir)
NONROOT_MODS := $(shell find . $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )

RECEIVER_MODS_0 := $(shell find ./receiver/[a-k]* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
RECEIVER_MODS_1 := $(shell find ./receiver/[l-z]* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
RECEIVER_MODS := $(RECEIVER_MODS_0) $(RECEIVER_MODS_1)
PROCESSOR_MODS := $(shell find ./processor/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
EXPORTER_MODS := $(shell find ./exporter/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
EXTENSION_MODS := $(shell find ./extension/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
INTERNAL_MODS := $(shell find ./internal/* $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) )
OTHER_MODS := $(shell find . $(EX_COMPONENTS) $(EX_INTERNAL) $(FIND_MOD_ARGS) -exec $(TO_MOD_DIR) ) $(PWD)
ALL_MODS := $(RECEIVER_MODS) $(PROCESSOR_MODS) $(EXPORTER_MODS) $(EXTENSION_MODS) $(INTERNAL_MODS) $(OTHER_MODS)

ifeq ($(GOOS),windows)
	EXTENSION := .exe
endif

.DEFAULT_GOAL := all

all-modules:
	@echo $(NONROOT_MODS) | tr ' ' '\n' | sort

all-groups:
	@echo "receiver-0: $(RECEIVER_MODS_0)"
	@echo "\nreceiver-1: $(RECEIVER_MODS_1)"
	@echo "\nreceiver: $(RECEIVER_MODS)"
	@echo "\nprocessor: $(PROCESSOR_MODS)"
	@echo "\nexporter: $(EXPORTER_MODS)"
	@echo "\nextension: $(EXTENSION_MODS)"
	@echo "\ninternal: $(INTERNAL_MODS)"
	@echo "\nother: $(OTHER_MODS)"

.PHONY: all
all: install-tools all-common goporto gotest nrotelcomponents

.PHONY: all-common
all-common:
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="common"

.PHONY: unit-tests-with-cover
unit-tests-with-cover: ## Runs unit tests with cover for the given $(FOR_GROUP_TARGET).
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="do-unit-tests-with-cover"

.PHONY: gotidy
gotidy: ## Runs go tidy for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="tidy"

.PHONY: gomoddownload
gomoddownload: ## Runs go mod download for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="moddownload"

.PHONY: gotest
gotest: ## Runs go test for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="test"

.PHONY: gofmt
gofmt: ## Runs gofmt for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="fmt"

.PHONY: golint
golint: ## Runs golangci-lint lint for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="lint"

.PHONY: goimpi
goimpi: install-tools ## Runs impi for the given $(FOR_GROUP_TARGET).
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="impi"

.PHONY: goporto
goporto: install-tools ## Runs porto tool for the given $(FOR_GROUP_TARGET).
	porto -w --include-internal --skip-dirs "^cmd$$" ./

.PHONY: gogenerate
gogenerate: install-tools ## Runs go generate for the given $(FOR_GROUP_TARGET).
	@$(MAKE) $(FOR_GROUP_TARGET) TARGET="generate"

.PHONY: for-all
for-all:
	@echo "running $${CMD} in root"
	@$${CMD}
	@set -e; for dir in $(NONROOT_MODS); do \
	  (cd "$${dir}" && \
	  	echo "running $${CMD} in $${dir}" && \
	 	$${CMD} ); \
	done

DEPENDABOT_PATH=".github/dependabot.yml"
.PHONY: gendependabot
gendependabot: ## Generates the dependabot configuration file.
	@echo "Recreating ${DEPENDABOT_PATH} file"
	@echo "# File generated by \"make gendependabot\"; DO NOT EDIT." > ${DEPENDABOT_PATH}
	@echo "" >> ${DEPENDABOT_PATH}
	@echo "version: 2" >> ${DEPENDABOT_PATH}
	@echo "updates:" >> ${DEPENDABOT_PATH}
	@echo "Add entry for \"/\" github-actions"
	@echo "  - package-ecosystem: \"github-actions\"" >> ${DEPENDABOT_PATH}
	@echo "    directory: \"/\"" >> ${DEPENDABOT_PATH}
	@echo "    schedule:" >> ${DEPENDABOT_PATH}
	@echo "      interval: \"weekly\"" >> ${DEPENDABOT_PATH}
	@echo "Add entry for \"/\" docker"
	@echo "  - package-ecosystem: \"docker\"" >> ${DEPENDABOT_PATH}
	@echo "    directory: \"/\"" >> ${DEPENDABOT_PATH}
	@echo "    schedule:" >> ${DEPENDABOT_PATH}
	@echo "      interval: \"weekly\"" >> ${DEPENDABOT_PATH}
	@echo "Add entry for \"/\" gomod"
	@echo "  - package-ecosystem: \"gomod\"" >> ${DEPENDABOT_PATH}
	@echo "    directory: \"/\"" >> ${DEPENDABOT_PATH}
	@echo "    schedule:" >> ${DEPENDABOT_PATH}
	@echo "      interval: \"weekly\"" >> ${DEPENDABOT_PATH}
	@set -e; for dir in $(NONROOT_MODS); do \
		echo "Add entry for \"$${dir:1}\""; \
		echo "  - package-ecosystem: \"gomod\"" >> ${DEPENDABOT_PATH}; \
		echo "    directory: \"$${dir:1}\"" >> ${DEPENDABOT_PATH}; \
		echo "    schedule:" >> ${DEPENDABOT_PATH}; \
		echo "      interval: \"weekly\"" >> ${DEPENDABOT_PATH}; \
	done

# Define a delegation target for each module
.PHONY: $(ALL_MODS)
$(ALL_MODS):
	@echo "Running target '$(TARGET)' in module '$@' as part of group '$(GROUP)'"
	$(MAKE) -C $@ $(TARGET)

# Trigger each module's delegation target
.PHONY: for-all-target
for-all-target: $(ALL_MODS)

.PHONY: for-receiver-target
for-receiver-target: $(RECEIVER_MODS)

.PHONY: for-receiver-0-target
for-receiver-0-target: $(RECEIVER_MODS_0)

.PHONY: for-receiver-1-target
for-receiver-1-target: $(RECEIVER_MODS_1)

.PHONY: for-processor-target
for-processor-target: $(PROCESSOR_MODS)

.PHONY: for-exporter-target
for-exporter-target: $(EXPORTER_MODS)

.PHONY: for-extension-target
for-extension-target: $(EXTENSION_MODS)

.PHONY: for-internal-target
for-internal-target: $(INTERNAL_MODS)

.PHONY: for-other-target
for-other-target: $(OTHER_MODS)

# Debugging target, which helps to quickly determine whether for-all-target is working or not.
.PHONY: all-pwd
all-pwd:
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="pwd"

TOOLS_MOD_DIR := ./internal/tools
.PHONY: install-tools
install-tools:
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/client9/misspell/cmd/misspell
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/google/addlicense
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/pavius/impi/cmd/impi
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/checkdoc
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install golang.org/x/tools/cmd/goimports
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/jcchavezs/porto/cmd/porto
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/crosslink

.PHONY: run
run:
	cd ./cmd/nrotelcomponents && GO111MODULE=on $(GOCMD) run --race . --config ../../${RUN_CONFIG} ${RUN_ARGS}

.PHONY: docker-component # Not intended to be used directly
docker-component: check-component
	GOOS=linux GOARCH=amd64 $(MAKE) $(COMPONENT)
	cp ./bin/$(COMPONENT)_linux_amd64 ./cmd/$(COMPONENT)/$(COMPONENT)
	docker build -t $(COMPONENT) ./cmd/$(COMPONENT)/
	rm ./cmd/$(COMPONENT)/$(COMPONENT)

.PHONY: check-component
check-component:
ifndef COMPONENT
	$(error COMPONENT variable was not defined)
endif

.PHONY: docker-nrotelcomponents
docker-nrotelcomponents: ## Build a docker image with the collector executable.
	COMPONENT=nrotelcomponents $(MAKE) docker-component


.PHONY: nrotelcomponents
nrotelcomponents: ## Build the collector executable.
	cd ./cmd/nrotelcomponents && GO111MODULE=on CGO_ENABLED=0 $(GOCMD) build -trimpath -o ../../bin/nrotelcomponents_$(GOOS)_$(GOARCH)$(EXTENSION) \
		$(BUILD_INFO) -tags $(GO_BUILD_TAGS) .

.PHONY: update-dep
update-dep: ## Update dependencies.
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="updatedep"
	$(MAKE) nrotelcomponents

.PHONY: update-otel
update-otel: ## Updates collector module version.
	$(MAKE) update-dep MODULE=go.opentelemetry.io/collector VERSION=$(OTEL_VERSION) RC_VERSION=$(OTEL_RC_VERSION) STABLE_VERSION=$(OTEL_STABLE_VERSION)

# Verify existence of READMEs for components specified as default components in the collector.
.PHONY: checkdoc
checkdoc: ## Checks that all components have documentation using checkdoc.
	checkdoc --project-path $(CURDIR) --component-rel-path $(COMP_REL_PATH) --module-name $(MOD_NAME)

.PHONY: all-checklinks
all-checklinks: ## Runs checklinks for the given $(FOR_GROUP_TARGET).
	$(MAKE) $(FOR_GROUP_TARGET) TARGET="checklinks"

# Function to execute a command. Note the empty line before endef to make sure each command
# gets executed separately instead of concatenated with previous one.
# Accepts command to execute as first parameter.
define exec-command
$(1)

endef

.PHONY: crosslink
crosslink: install-tools ## Runs crosslink tool.
	@echo "Executing crosslink"
	crosslink --root=$(shell pwd)


.PHONY: clean
clean: ## Removes coverage files
	@echo "Removing coverage files"
	find . -type f -name 'coverage.txt' -delete
	find . -type f -name 'coverage.html' -delete
	find . -type f -name 'integration-coverage.txt' -delete
	find . -type f -name 'integration-coverage.html' -delete
	find . -type f -name 'foresight-test-report.txt' -delete
