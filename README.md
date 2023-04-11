# Conch: The Conventional Commits Checker

## Features

Conch verifies that your Git commit messages adhere to the
[Conventional Commits] specification.
That is, all commit messages must use the following format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Conventional Commits encourages developers to write better, more descriptive
commit messages. It also enables commit messages to be machine-parseable.
Use Conch as a pull request check to enforce these best practices.

Conch can also inspect your commits and filter them by Conventional Commit attributes
like type and scope. Use Conch to prepare changelogs, summarize breaking changes,
and automate other CI workflows.

## Quick Start

### Standalone Version

Download Conch from the releases page, place it somewhere on your path
(like `/usr/local/bin/`), and ensure the file is executable.

Then, run the program from within your Git repository.

```bash
conch 'HEAD~10..'
```

The `conch` binary is statically linked, so no dependencies are needed.

### Docker Version

Conch is available on Docker Hub. See the repository page for supported image tags
and architectures. (Please pin to a specific release branch instead of `latest`.)

```bash
docker pull csang/conch:0.1
```

When running the container, bind mount your repository and set it as the container's
working directory. (A read-only mount is sufficient.) For example:

```bash
docker run --rm -v "$(pwd):/mnt/workspace:ro" --workdir=/mnt/workspace csang/conch:0.1 'HEAD~10..'
```

### Github Actions

_coming soon_

### Pre-Commit Hook

_coming soon_

### Go Module

_coming soon_

## Full Usage Instructions

```
Usage: conch [options] <revision_range>
  -h, --help                             display this help text
  -q, --quiet                            suppress error messages for bad commits
  -v, --verbose                          verbose log output
  -V, --version                          display version and build info
  -c, --config string                    path to config file
  -r, --repo string                      path to the git repository
  -T, --types comma_separated_strings    filter commits by type
  -S, --scopes comma_separated_strings   filter commits by scope
  -B, --breaking                         show breaking changes (e.g., feat!)
  -M, --minor                            show minor changes (e.g., feat)
  -P, --patch                            show patch changes (e.g., fix)
  -U, --uncategorized                    show other changes that are not breaking/minor/patch
  -l, --list                             list matching commits
  -f, --format string                    format matching commits using a Go template
  -n, --count                            show the number of matching commits
  -i, --impact                           show the max impact of the commits (breaking/minor/patch/uncategorized)
  -b, --bump-version string              bump up the specified version number based on the changes in the range
```

### Revision Range

`conch` requires one positional argument specifying the range of commits to parse.
For example:

```bash
# Validate the ten most recent commits on the current branch
conch 'HEAD~10..'

# Validate all the commits on branch "dev"
conch 'main..dev'

# Validate a specific range of commit hashes
conch '40b9741..2453f95'
```

See the [Git documentation](https://git-scm.com/book/en/v2/Git-Tools-Revision-Selection)
for more tips on how to specify a commit range.

### Git Repository Location

In most cases, you should run `conch` from within your project's working directory,
just as you would run other `git` commands. Use the `--repo` flag if you need to point
`conch` at a different directory. For Docker, you can also set the working directory
as part of the run command, `docker run --workdir`.

### Output Options

`conch` validates the range of commits and reports any that violate
the Conventional Commits standard.

Add an output option to display more information about the commits that were
successfully parsed.

#### List Commits (`-l`, `--list`)

```bash
conch -l 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
46597ca: feat: add issue reporting links
647e997: chore(deps): upgrade gems
40d1d41: feat(post): alpine linux
36a3e9d: feat(post): python type annotations
```

Commits are listed in a human-readable format. Use a format specifier
if you need to generate custom machine-readable output.

#### Format Commits (`-f`, `--format`)

```bash
conch -f '{{ .Id }} {{ .Type }}\n' 'HEAD~5..'
```

```
2453f95585b93dc14bb986191e422c31e76171b4 fix
46597ca0f173b7675e4aa26d5fdee30e49be83eb feat
647e99725a12f107f73bb835b185fb73fc19e5d8 chore
40d1d41194a5b9b08945c8778b7a7fc55046c749 feat
36a3e9d2ef0e5157952650f50d2613c5dc079748 feat
```

The format specifier is a [Go template](https://pkg.go.dev/text/template)
which accepts the following variables:

```ini
.Id           # The git commit hash
.Type         # The commit type
.Scope        # The commit scope (may be empty)
.Description  # The commit description (may be empty)
.Body         # The remainder of the commit message, excluding any footers (may be empty)
.Footers      # The footers, as a list of {Token, Value} pairs (may be empty)
```

You may also use the following escape sequences:

* `\t` - tab
* `\n` - newline
* `\\` - literal backslash

#### Count Commits (`-n`, `--count`)

```bash
conch -n 'HEAD~5..'
```

```
5
```

#### Determine Impact of Changes (`-i`, `--impact`)

Given the commits in the range, show the highest impact of the changes
based on the commit types and any breaking change designations:

* `breaking`
* `minor`
* `patch`
* `uncategorized`

For example, `conch -i 'HEAD~5'` returns `breaking` if at least one of
the previous five commits contains a breaking change. It returns `minor`
if there is at least one `fix`, and no breaking changes.

#### Bump Up the Version Number (`-b`, `--bump-version`)

Given the specified version number, output the next version number
in the sequence based on the impact of the commits in the range.
Use this function to help automate tagged releases of your project.

For example, `conch -b '1.0.0' 'HEAD~5'` returns `2.0.0` if at least one of
the previous five commits contains a breaking change. It returns `1.1.0` if
there is a `fix` or other minor commit in the range, and no breaking changes.

The starting version number must be a [semantic version][semver]:

```
major.minor.patch[-prerelease_info][+build_metadata]
```

```
1.2.3
1.2.3-alpha.1+build.92690d
```

Note: Prerelease info and build metadata is always stripped from the output.
Major version zero (often used during initial development) is not treated
specially.

### Filter Options

Use a filter option to control the output.

* Filters do not affect commit validation. All commits in the range
  are validated against the Conventional Commits specification, regardless
  of filter options.
* In accordance with the specification, all filters perform case-insensitive
  matching, except for `-B`, `--breaking`.

#### Types (`-T`, `--types`)

```bash
conch -T 'feat,fix' 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
46597ca: feat: add issue reporting links
40d1d41: feat(post): alpine linux
36a3e9d: feat(post): python type annotations
```

#### Scopes (`-S`, `--scopes`)

```bash
conch -S post 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
40d1d41: feat(post): alpine linux
36a3e9d: feat(post): python type annotations
```

Commits without a scope:

```bash
conch -S '' 'HEAD~5..'
```

```
46597ca: feat: add issue reporting links
```

Commits with scope "post", or commits without a scope:

```bash
conch -S 'post,' 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
46597ca: feat: add issue reporting links
40d1d41: feat(post): alpine linux
36a3e9d: feat(post): python type annotations
```

#### Impact

* `-B`, `--breaking`: select commits marked with `!` or a `BREAKING CHANGE` footer.
* `-M`, `--minor`: select minor changes like `feat`
* `-P`, `--patch`: select patch changes like `fix`
* `-U`, `--uncategorized`: select all other commit types

Example: List all patches and other low-impact changes:

```bash
conch -P -U 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
647e997: chore(deps): upgrade gems
```

To customize which commit types are treated as minor and patch, use a `conch.yml`
configuration file, described later in this document.

#### Multiple Filter Options

A commit matches the filters if the type AND scope are correct, AND the impact
of the change matches one of the impact filters.

```bash
conch -T fix -S post -M -P -U 'HEAD~5..'
```

```
2453f95: fix(post): add runServices to dev container sample code
```

### Exit Status

Conch exits successfully if all commits in the range comply with the
Conventional Commits specification. Otherwise, it exits with a non-zero
status code.

## Configuration File

Conch can enforce custom commit policies. Example scenarios:

* Require a specific set of commit types, scopes, or footers
* Require all commits to specify a scope
* Limit the length of the commit description
* Ignore certain commit message patterns

To customize the behavior of Conch, create a `conch.yml` file at the root
of your repository. Use the [`conch.default.yml`](conch.default.yml) file
as a starting point for your configuration, and see the comments there
explaining the file format.

Note: If you need to put your configuration file somewhere else, you can
select it via `-c` or `--config`:

```bash
conch -c '/alternate/path/to/conch.yml' 'HEAD~5..'
```

## Developer Information

Example run command:

```bash
docker-compose run --build --rm app
```

Available services:

* `app` - contains the compiled binaries, ready for production deployment
* `dev` - dev container for running arbitrary build and debug commands
* `test` - run tests in dev container
* `cover` - run tests with code coverage

## License

Conch is licensed under the MIT license. See `LICENSE.txt`.

Conch is statically-linked against the following third-party libraries:

* libc (musl) ([MIT](https://git.musl-libc.org/cgit/musl/tree/COPYRIGHT))
* libgit2 ([Linking Exception to GPL v2](https://github.com/libgit2/libgit2/blob/main/COPYING))
* libhttp_parser ([MIT](https://github.com/nodejs/http-parser/blob/main/LICENSE-MIT))
* libssh2 ([BSD-3](https://github.com/libssh2/libssh2/blob/master/COPYING))
* libssl (openssl) ([Apache 2.0](https://github.com/openssl/openssl/blob/master/LICENSE.txt))
* zlib ([License](https://github.com/madler/zlib/blob/master/LICENSE))

It references the following standards:

* [Conventional Commits] ([CC-BY-3.0])
* [Semantic Versioning][semver] ([CC-BY-3.0])

[Conventional Commits]: https://www.conventionalcommits.org/
[semver]: https://semver.org/
[CC-BY-3.0]: https://creativecommons.org/licenses/by/3.0/
