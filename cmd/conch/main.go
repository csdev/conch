package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/csdev/conch/internal/commit"
	"github.com/csdev/conch/internal/util"
	flag "github.com/spf13/pflag"
)

func enforceExclusiveFlags(groupName string, flagNames ...string) error {
	var changed bool
	for _, f := range flagNames {
		if flag.CommandLine.Changed(f) {
			if changed {
				names := strings.Join(flagNames, ", ")
				return fmt.Errorf("%s (%s) are mutually exclusive", groupName, names)
			} else {
				changed = true
			}
		}
	}
	return nil
}

func main() {
	var (
		help    bool
		verbose bool
		version bool

		configPath string
		repoPath   string

		filterTypes    util.CaseInsensitiveSet
		filterScopes   util.CaseInsensitiveSet
		filterBreaking bool
		filterMinor    bool
		filterPatch    bool
		filterOther    bool

		outputList   bool
		outputFormat string
		outputCount  bool
	)

	// meta
	flag.BoolVarP(&help, "help", "h", help, "display this help text")
	flag.BoolVarP(&verbose, "verbose", "v", verbose, "verbose log output")
	flag.BoolVarP(&version, "version", "V", version, "display version and build info")

	// configuration
	flag.StringVarP(&configPath, "config", "c", configPath, "path to config file")
	flag.StringVarP(&repoPath, "repo", "r", repoPath, "path to the git repository")

	// output filtering
	flag.VarP(&filterTypes, "types", "T", "filter commits by type")
	flag.VarP(&filterScopes, "scopes", "S", "filter commits by scope")
	flag.BoolVarP(&filterBreaking, "breaking", "B", filterBreaking, "show breaking changes (e.g., feat!)")
	flag.BoolVarP(&filterMinor, "minor", "M", filterMinor, "show minor changes (e.g., feat)")
	flag.BoolVarP(&filterPatch, "patch", "P", filterPatch, "show patch changes (e.g., fix)")
	flag.BoolVarP(&filterOther, "uncategorized", "U", filterOther,
		"show other changes that are not breaking/minor/patch")

	// output formatting
	flag.BoolVarP(&outputList, "list", "l", outputList, "list matching commits")
	flag.StringVarP(&outputFormat, "format", "f", outputFormat,
		"format matching commits using a Go template")
	flag.BoolVarP(&outputCount, "count", "n", outputCount, "show the number of matching commits")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		// HACK: Zero out custom `VarP` flags, or else they cause blank
		// help text for default values to be added to the output.
		// https://github.com/spf13/pflag/issues/245
		// When calling Usage(), the program should exit soon after,
		// so doing this shouldn't actually break normal operation.
		filterTypes = nil
		filterScopes = nil

		fmt.Fprintf(os.Stderr, "Usage: %s [options] <revision_range>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if help {
		flag.Usage()
		return
	}
	if version {
		fmt.Fprintln(os.Stderr, "conch")
		bi, ok := debug.ReadBuildInfo()
		if ok {
			fmt.Fprintf(os.Stderr, "+%v\n", bi)
		} else {
			fmt.Fprintln(os.Stderr, "build information is not available")
		}
		return
	}

	if err := enforceExclusiveFlags("output flags", "list", "format", "count"); err != nil {
		flag.Usage()
		log.Fatalf("error: %v\n", err)
	}

	if flag.NArg() != 1 {
		flag.Usage()
		log.Fatalln("error: please specify a revision range")
	}

	if repoPath == "" {
		repoPath = os.Getenv("GITHUB_WORKSPACE")
	}
	if repoPath == "" {
		repoPath = os.Getenv("CONCH_DOCKER_WORKSPACE")
	}
	if repoPath == "" {
		repoPath = "."
	}

	var tpl *template.Template
	if outputFormat != "" {
		repFormat := strings.NewReplacer(`\t`, "\t", `\n`, "\n").Replace(outputFormat)
		var err error
		tpl, err = template.New("commit").Parse(repFormat)
		if err != nil {
			log.Fatalf("invalid template: %v", err)
		}
	}

	commits, parseErr := commit.ParseRange(repoPath, flag.Arg(0))
	if parseErr != nil {
		log.Printf("%v", parseErr)
		// don't exit yet -- try outputting any valid commits that were found
	}

	var numCommits int

	needsOutput := (filterScopes != nil || filterTypes != nil ||
		filterBreaking || filterMinor || filterPatch || filterOther ||
		outputList || outputFormat != "" || outputCount)

	if needsOutput {
		for _, c := range commits {
			if filterTypes != nil && !filterTypes.Contains(c.Type) {
				continue
			}
			if filterScopes != nil && !filterScopes.Contains(c.Scope) {
				continue
			}

			if tpl != nil {
				tpl.Execute(os.Stdout, c)
			} else if !outputCount {
				fmt.Printf("%s: %s\n", c.Id[:7], c.Summary())
			}
			numCommits += 1
		}
	}

	if outputCount {
		fmt.Printf("%d\n", numCommits)
	}

	if parseErr != nil {
		log.Fatalln("failed to parse some commits")
	}
}
