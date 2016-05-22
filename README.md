# Updater

[![Build Status](https://travis-ci.org/keybase/go-updater.svg?branch=master)](https://travis-ci.org/keybase/go-updater)
[![Build Status](https://ci.appveyor.com/api/projects/status/github/keybase/go-updaterbranch=master&svg=true)](https://ci.appveyor.com/project/keybase/go-updater)
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


### Development

Install the pre-commit hook (and run manually):

```
pip install pre-commit
go get -u github.com/alecthomas/gometalinter
gometalinter --install --update
pre-commit install
pre-commit run -a
```
