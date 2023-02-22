package util

import (
	"strings"
	"text/template"
)

func OutputTemplate(name string, format string) (*template.Template, error) {
	repFormat := strings.NewReplacer(`\t`, "\t", `\n`, "\n").Replace(format)
	return template.New("commit").Parse(repFormat)
}
