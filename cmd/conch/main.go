package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/csdev/conch/internal/cli"
	"github.com/csdev/conch/internal/commit"
	"github.com/csdev/conch/internal/config"
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

		filters cli.Filters
		outputs cli.Outputs
	)

	// meta
	flag.BoolVarP(&help, "help", "h", help, "display this help text")
	flag.BoolVarP(&verbose, "verbose", "v", verbose, "verbose log output")
	flag.BoolVarP(&version, "version", "V", version, "display version and build info")

	// configuration
	flag.StringVarP(&configPath, "config", "c", configPath, "path to config file")
	flag.StringVarP(&repoPath, "repo", "r", repoPath, "path to the git repository")

	// output filtering
	flag.VarP(&filters.Types, "types", "T", "filter commits by type")
	flag.VarP(&filters.Scopes, "scopes", "S", "filter commits by scope")

	flag.BoolVarP(&filters.Selections.Breaking, "breaking", "B", filters.Selections.Breaking,
		"show breaking changes (e.g., feat!)")
	flag.BoolVarP(&filters.Selections.Minor, "minor", "M", filters.Selections.Minor,
		"show minor changes (e.g., feat)")
	flag.BoolVarP(&filters.Selections.Patch, "patch", "P", filters.Selections.Patch,
		"show patch changes (e.g., fix)")
	flag.BoolVarP(&filters.Selections.Uncategorized, "uncategorized", "U", filters.Selections.Uncategorized,
		"show other changes that are not breaking/minor/patch")

	// output formatting
	flag.BoolVarP(&outputs.List, "list", "l", outputs.List,
		"list matching commits")
	flag.StringVarP(&outputs.Format, "format", "f", outputs.Format,
		"format matching commits using a Go template")
	flag.BoolVarP(&outputs.Count, "count", "n", outputs.Count,
		"show the number of matching commits")

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		// HACK: Zero out custom `VarP` flags, or else they cause blank
		// help text for default values to be added to the output.
		// https://github.com/spf13/pflag/issues/245
		// When calling Usage(), the program should exit soon after,
		// so doing this shouldn't actually break normal operation.
		filters.Types = nil
		filters.Scopes = nil

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
	if outputs.Format != "" {
		var err error
		tpl, err = util.OutputTemplate("commit", outputs.Format)
		if err != nil {
			log.Fatalf("invalid template: %v", err)
		}
	}

	cfg := config.Default()
	if configPath == "" {
		p, err := cli.StandardConfigPath(repoPath)
		if err != nil {
			log.Fatalf("config error: %v", err)
		}
		configPath = p
	}
	if configPath != "" {
		// open specified config file
		file, err := os.Open(configPath)
		if err != nil {
			log.Fatalf("config error: %v", err)
		}
		cfg, err = config.Load(file)
		if err != nil {
			log.Fatalf("config error: %v", err)
		}
	}

	commits, parseErr := commit.ParseRange(repoPath, flag.Arg(0))
	if parseErr != nil {
		log.Printf("%v", parseErr)
		// don't exit yet -- try outputting any valid commits that were found
	}

	var numCommits int
	selectAll := !filters.Selections.Any()

	if filters.Any() || outputs.Any() {
		for _, c := range commits {
			if filters.Types != nil && !filters.Types.Contains(c.Type) {
				continue
			}
			if filters.Scopes != nil && !filters.Scopes.Contains(c.Scope) {
				continue
			}

			selected := selectAll
			if filters.Selections.Breaking && c.IsBreaking {
				selected = true
			}
			if filters.Selections.Minor && cfg.Minor.Contains(c.Type) {
				selected = true
			}
			if filters.Selections.Patch && cfg.Patch.Contains(c.Type) {
				selected = true
			}

			if !selected {
				continue
			}

			if tpl != nil {
				err := tpl.Execute(os.Stdout, c)
				if err != nil {
					log.Printf("%v", err)
				}
			} else if !outputs.Count {
				fmt.Printf("%s: %s\n", c.Id[:7], c.Summary())
			}
			numCommits += 1
		}
	}

	if outputs.Count {
		fmt.Printf("%d\n", numCommits)
	}

	if parseErr != nil {
		log.Fatalln("failed to parse some commits")
	}
}
