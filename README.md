# Updater

[![Build Status](https://travis-ci.org/keybase/go-updater.svg?branch=master)](https://travis-ci.org/keybase/go-updater)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/keybase/go-updater?branch=master&svg=true)](https://ci.appveyor.com/project/keybase/go-updater)
[![Coverage Status](https://coveralls.io/repos/github/keybase/go-updater/badge.svg?branch=master)](https://coveralls.io/github/keybase/go-updater?branch=master)
[![GoDoc](https://godoc.org/github.com/keybase/go-updater?status.svg)](https://godoc.org/github.com/keybase/go-updater)

**Warning**: This isn't ready for non-Keybase libraries to use yet!

The goals of this library are to provide an updater that:

- Is simple
- Works on all our platforms (at least OS X, Windows, Linux)
- Recovers from non-fatal errors
- Can recover from failures in its environment
- Can run as an unpriviledged background service
- Has minimal, vendored dependencies
- Is well tested
- Is secure
- Reports failures and activity
- Can notify the user of any non-transient failures

This updater library is used to support updating (in background and on-demand)
for Keybase apps and services.


### Packages

- command: Executes a command with a timeout
- keybase: Keybase specific behavior for updates
- osx: Presents UI for OS X
- process: Utilities to find and terminate Processes
- service: Runs the updater as a background service
- sources: Update sources (locations) for static remote (like S3), or locally (for testing)
- test: Test resources
- util: Utilities for updating, such as digests, env, file, http, unzip, etc.


### Development

This libray should pass the [gometalinter](https://github.com/alecthomas/gometalinter).

There is a pre-commit hook available:

```
pip install pre-commit
go get -u github.com/alecthomas/gometalinter
gometalinter --install --update
pre-commit install
pre-commit run -a
```
