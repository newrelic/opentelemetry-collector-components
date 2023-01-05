//go:build tools

package tools

import (
	// Testing
	_ "github.com/golang/mock/mockgen"
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/stretchr/testify/assert"
	_ "gotest.tools/gotestsum"

	// Linting
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/llorllale/go-gitlint/cmd/go-gitlint"
	_ "github.com/psampaz/go-mod-outdated"
	_ "golang.org/x/tools/cmd/goimports"

	// Documentation
	_ "golang.org/x/tools/cmd/godoc"

	// Releasing
	_ "github.com/caarlos0/svu"
	_ "github.com/git-chglog/git-chglog/cmd/git-chglog"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/x-motemen/gobump/cmd/gobump"

	// Generating
	_ "golang.org/x/tools/cmd/stringer"
)
