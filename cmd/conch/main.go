package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/csdev/conch/internal/cli"
	"github.com/csdev/conch/internal/commit"
	"github.com/csdev/conch/internal/config"
	"github.com/csdev/conch/internal/semver"
	log "github.com/sirupsen/logrus"
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

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})
}

func main() {
	var (
		help    bool
		quiet   bool
		verbose bool
		version bool

		configPath string
		repoPath   string

		hook bool

		filters cli.Filters
		outputs cli.Outputs
	)

	// meta
	flag.BoolVarP(&help, "help", "h", help, "display this help text")
	flag.BoolVarP(&quiet, "quiet", "q", quiet, "suppress error messages for bad commits")
	flag.BoolVarP(&verbose, "verbose", "v", verbose, "verbose log output")
	flag.BoolVarP(&version, "version", "V", version, "display version and build info")

	// configuration
	flag.StringVarP(&configPath, "config", "c", configPath, "path to config file")
	flag.StringVarP(&repoPath, "repo", "r", repoPath, "path to the git repository")

	// git hook mode
	flag.BoolVarP(&hook, "hook", "k", hook, "run as git commit-msg hook, validating a file (see docs)")

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
	flag.BoolVarP(&outputs.Impact, "impact", "i", outputs.Impact,
		"show the max impact of the commits (breaking/minor/patch/uncategorized)")
	flag.StringVarP(&outputs.BumpVersion, "bump-version", "b", outputs.BumpVersion,
		"bump up the specified version number based on the changes in the range")

	flagGroups := map[string][]string{
		"log options": {
			"quiet",
			"verbose",
		},
		"output flags": {
			"list",
			"format",
			"count",
			"impact",
			"bump-version",
		},
	}

	flag.CommandLine.SortFlags = false

	flag.Usage = func() {
		// HACK: Zero out custom `VarP` flags, or else they cause blank
		// help text for default values to be added to the output.
		// https://github.com/spf13/pflag/issues/245
		// When calling Usage(), the program should exit soon after,
		// so doing this shouldn't actually break normal operation.
		filters.Types = nil
		filters.Scopes = nil

		const usage = "Usage: %s [options] <revision_range>\n" +
			"       %s [-k|--hook] <filename>\n"

		fmt.Fprintf(os.Stderr, usage, os.Args[0], os.Args[0])
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

	for groupName, flagNames := range flagGroups {
		if err := enforceExclusiveFlags(groupName, flagNames...); err != nil {
			flag.Usage()
			log.Fatalf("%v", err)
		}
	}

	if flag.NArg() != 1 {
		flag.Usage()
		if hook {
			log.Fatalln("commit-msg hook: please specify a filename")
		} else {
			log.Fatalln("please specify a revision range")
		}
	}

	if quiet {
		log.SetLevel(log.FatalLevel)
	} else if verbose {
		log.SetLevel(log.DebugLevel)
	}

	var sv *semver.Semver
	if outputs.BumpVersion != "" {
		var err error
		sv, err = semver.Parse(outputs.BumpVersion)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	if repoPath == "" {
		repoPath = "."
	}

	var tpl *template.Template
	if outputs.Format != "" {
		var err error
		tpl, err = cli.Template("commit", outputs.Format)
		if err != nil {
			log.Fatalf("invalid template: %v", err)
		}
	}

	if configPath == "" {
		p, err := config.Discover(repoPath)
		if err != nil {
			log.Fatalf("config: %v", err)
		}
		configPath = p
	}
	cfg, err := config.Open(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	var origMsg string
	var commits []*commit.Commit
	var parseErr error

	if hook {
		origMsg, parseErr = cli.GetFileContents(flag.Arg(0))
		if parseErr != nil {
			log.Fatalf("%v", parseErr)
		}
		origMsg = commit.StripComments(origMsg)
		commits, parseErr = commit.ParseMessage(origMsg, cfg)
	} else {
		commits, parseErr = commit.ParseRange(repoPath, flag.Arg(0), cfg)
	}

	if parseErr != nil {
		log.Errorf("%v", parseErr)
		// don't exit yet -- try outputting any valid commits that were found
	}

	policyErr := commit.ApplyPolicy(commits, cfg)
	if policyErr != nil {
		log.Errorf("%v", policyErr)
		// don't exit yet -- try outputting any valid commits that were found
	}

	var numCommits int
	impact := commit.Uncategorized
	selectAll := !filters.Selections.Any()

	if filters.Any() && !outputs.Any() {
		outputs.List = true
	}

	if outputs.Any() {
		for _, c := range commits {
			if filters.Types != nil && !filters.Types.Contains(c.Type) {
				continue
			}
			if filters.Scopes != nil && !filters.Scopes.Contains(c.Scope) {
				continue
			}

			cls := c.Classification(cfg)
			selected := selectAll

			if filters.Selections.Breaking && cls == commit.Breaking {
				selected = true
			}
			if filters.Selections.Minor && cls == commit.Minor {
				selected = true
			}
			if filters.Selections.Patch && cls == commit.Patch {
				selected = true
			}
			if filters.Selections.Uncategorized && cls == commit.Uncategorized {
				selected = true
			}

			if !selected {
				continue
			}

			if tpl != nil {
				err := tpl.Execute(os.Stdout, c)
				if err != nil {
					log.Errorf("%v", err)
				}
			} else if outputs.List {
				fmt.Printf("%s: %s\n", c.ShortId, c.Summary())
			}
			numCommits += 1

			if cls < impact {
				impact = cls
			}
		}
	}

	if outputs.Count {
		fmt.Printf("%d\n", numCommits)
	} else if outputs.Impact {
		fmt.Printf("%s\n", []string{"breaking", "minor", "patch", "uncategorized"}[impact])
	} else if sv != nil {
		var nextVer *semver.Semver
		switch impact {
		case commit.Breaking:
			nextVer = sv.NextMajor()
		case commit.Minor:
			nextVer = sv.NextMinor()
		case commit.Patch:
			nextVer = sv.NextPatch()
		default:
			nextVer = sv.NextRelease()
		}
		fmt.Printf("%s\n", nextVer.String())
	}

	if parseErr != nil || policyErr != nil {
		if quiet {
			os.Exit(1)
		} else {
			if origMsg != "" {
				fmt.Fprintf(os.Stderr, "original commit message:\n%s\n", origMsg)
			}
			log.Fatalln("failed to parse some commits")
		}
	}
}
