#
# Makefile fragment for Linting
#

GO           ?= go
MISSPELL     ?= $(GOBIN)/misspell
GOFMT        ?= gofmt
GOIMPORTS    ?= $(GOBIN)/goimports

COMMIT_LINT_CMD   ?= $(GOBIN)/go-gitlint
COMMIT_LINT_REGEX ?= "(chore|docs|feat|fix|refactor|tests?)\s?(\([^\)]+\))?!?: .*"
COMMIT_LINT_START ?= "2021-08-17"

GOLINTER      = $(GOBIN)/golangci-lint

EXCLUDEDIR      ?= .git
SRCDIR          ?= .
GO_PKGS         ?= $(shell ${GO} list ./... | grep -v -e "/vendor/" -e "/example")
FILES           ?= $(shell find ${SRCDIR} -type f | grep -v -e '.git/' -e '/vendor/' -e 'go.\(mod\|sum\)' -e 'bin/' -e '.golangci.yml')
GO_FILES        ?= $(shell find $(SRCDIR) -type f -name "*.go" | grep -v -e ".git/" -e '/vendor/' -e '/example/')
PROJECT_MODULE  ?= $(shell $(GO) list -m)

GO_MOD_OUTDATED ?= $(GOBIN)/go-mod-outdated

lint: deps spell-check gofmt lint-commit golangci goimports outdated
lint-fix: deps spell-check-fix gofmt-fix goimports

#
# Check spelling on all the files, not just source code
#
spell-check: deps
	@echo "=== $(PROJECT_NAME) === [ spell-check      ]: Checking for spelling mistakes..."
	@$(MISSPELL) -source text $(FILES)

spell-check-fix: deps
	@echo "=== $(PROJECT_NAME) === [ spell-check-fix  ]: Fixing spelling mistakes..."
	@$(MISSPELL) -source text -w $(FILES)

gofmt: deps
	@echo "=== $(PROJECT_NAME) === [ gofmt            ]: Checking file format..."
	@find . -path "$(EXCLUDEDIR)" -prune -print0 | xargs -0 $(GOFMT) -e -l -s -d ${SRCDIR}

gofmt-fix: deps
	@echo "=== $(PROJECT_NAME) === [ gofmt-fix        ]: Fixing file format..."
	@find . -path "$(EXCLUDEDIR)" -prune -print0 | xargs -0 $(GOFMT) -e -l -s -w ${SRCDIR}

goimports: deps
	@echo "=== $(PROJECT_NAME) === [ goimports        ]: Checking imports..."
	@$(GOIMPORTS) -w -local $(PROJECT_MODULE) $(GO_FILES)

lint-commit: deps
	@echo "=== $(PROJECT_NAME) === [ lint-commit      ]: Checking that commit messages are properly formatted..."
	@$(COMMIT_LINT_CMD) --since=$(COMMIT_LINT_START) --subject-minlen=10 --subject-maxlen=150 --subject-regex=$(COMMIT_LINT_REGEX)

golangci: deps
	@echo "=== $(PROJECT_NAME) === [ golangci-lint    ]: Linting..."
	@$(GOLINTER) run

outdated: deps tools-outdated
	@echo "=== $(PROJECT_NAME) === [ outdated         ]: Finding outdated deps..."
	@$(GO) list -u -m -json all | $(GO_MOD_OUTDATED) -direct -update

.PHONY: lint spell-check spell-check-fix gofmt gofmt-fix lint-fix lint-commit outdated goimports
