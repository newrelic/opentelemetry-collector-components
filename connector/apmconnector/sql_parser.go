// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmconnector // import "github.com/newrelic/opentelemetry-collector-components/connector/apmconnector"

import (
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

var re = regexp.MustCompile(`(?i).*?\sfrom[\s\[]+([^\]\s,)(;]*).*`)

type SQLParser struct {
}

func NewSQLParser() *SQLParser {
	return &SQLParser{}
}

func (sqlParser *SQLParser) ParseDbTableFromSQL(sql string) (string, bool) {
	matches := re.FindStringSubmatch(sql)
	count := len(matches)
	if count < 2 {
		return "", false
	}
	return strings.ToLower(matches[1]), true
}

func (sqlParser *SQLParser) ParseDbTableFromSpan(span ptrace.Span) (string, bool) {
	dbTable, dbTablePresent := span.Attributes().Get(DbSQLTableAttributeName)
	if dbTablePresent {
		return dbTable.AsString(), false
	}
	if sql, sqlPresent := span.Attributes().Get("db.statement"); sqlPresent {
		if parsedTable, exists := sqlParser.ParseDbTableFromSQL(sql.AsString()); exists {
			return parsedTable, true
		}
	}
	return "unknown", false
}
