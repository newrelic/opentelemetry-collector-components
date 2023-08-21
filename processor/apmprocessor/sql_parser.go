// Copyright New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apmprocessor // import "apmprocessor"

import (
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	DbSQLTableAttributeName = "db.sql.table"
	DbStatement             = "db.statement"
)

type SQLParser struct {
	re *regexp.Regexp
}

func NewSQLParser() *SQLParser {
	re, _ := regexp.Compile(`(?i).*?\sfrom[\s\[]+([^\]\s,)(;]*).*`)
	return &SQLParser{re: re}
}

func (sqlParser *SQLParser) ParseDbTableFromSQL(sql string) (string, bool) {
	matches := sqlParser.re.FindStringSubmatch(sql)
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
	if sql, sqlPresent := span.Attributes().Get(DbStatement); sqlPresent {
		if parsedTable, exists := sqlParser.ParseDbTableFromSQL(sql.AsString()); exists {
			return parsedTable, true
		}
	}
	return "unknown", false
}
