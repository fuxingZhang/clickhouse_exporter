package util

import (
	"regexp"
	"strings"
)

var regSQLFormat = regexp.MustCompile(`\s+`)

// FormatSQL format sql
func FormatSQL(sql string) string {
	return strings.TrimSpace(regSQLFormat.ReplaceAllString(sql, " "))
}
