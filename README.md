# Conch: The Conventional Commits Checker

## Features

Conch verifies that your Git commit messages adhere to the
[Conventional Commits](https://www.conventionalcommits.org/) specification.
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

When running the container, bind mount your repository to `/mnt/workspace`:

```bash
docker run --rm -v "/path/to/git/repository:/mnt/workspace" csang/conch:0.1 'HEAD~10..'
```

For a repository in the current working directory:

```bash
docker run --rm -v "$(pwd):/mnt/workspace" csang/conch:0.1 'HEAD~10..'
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

Conch looks for your Git repository in the following locations:

1. The path specified in the `--repo` flag.
2. The `GITHUB_WORKSPACE` environment variable (for use with Github Actions).
3. The bind-mounted directory `/mnt/workspace`, when running in Docker.
4. The current working directory.

In most cases, you should run `conch` from within your project's working directory,
just as you would run other `git` commands. Use the `--repo` flag if you need to point
`conch` at a different directory.

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

Additionally, you may use `\t` and `\n` to insert tabs and newlines.

#### Count Commits (`-n`, `--count`)

```bash
conch -n 'HEAD~5..'
```

```
5
```

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

#### Multiple Filter Options

Filters use AND logic:

```bash
conch -T fix -S post 'HEAD~5..'
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
