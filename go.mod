module github.com/keybase/go-updater

go 1.17

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/kardianos/osext v0.0.0-20150528142315-6e7f84366347
	github.com/keybase/client/go v0.0.0-20211119210509-040230869410
	github.com/keybase/go-logging v0.0.0-20211118164508-35a15a9fa71a
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/keybase/saltpack v0.0.0-20211118165207-4039c5df46c0
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20211117180635-dee7805ff2e1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/keybase/backoff v1.0.1-0.20160517061000-726b63b835ec // indirect
	github.com/keybase/clockwork v0.1.1-0.20161209210251-976f45f4a979 // indirect
	github.com/keybase/go-codec v0.0.0-20180928230036-164397562123 // indirect
	github.com/keybase/go-crypto v0.0.0-20200123153347-de78d2cb44f4 // indirect
	github.com/keybase/go-framed-msgpack-rpc v0.0.0-20211118173254-f892386581e8 // indirect
	github.com/keybase/go-jsonw v0.0.0-20200325173637-df90f282c233 // indirect
	github.com/keybase/msgpackzip v0.0.0-20211109205514-10e4bc329851 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// keybase maintained forks
replace (
	bazil.org/fuse => github.com/keybase/fuse v0.0.0-20210104232444-d36009698767
	bitbucket.org/ww/goautoneg => github.com/adjust/goautoneg v0.0.0-20150426214442-d788f35a0315
	github.com/stellar/go => github.com/keybase/stellar-org v0.0.0-20191010205648-0fc3bfe3dfa7
	github.com/syndtr/goleveldb => github.com/keybase/goleveldb v1.0.1-0.20211106225230-2a53fac0721c
	gopkg.in/src-d/go-billy.v4 => github.com/keybase/go-billy v3.1.1-0.20180828145748-b5a7b7bc2074+incompatible
	gopkg.in/src-d/go-git.v4 => github.com/keybase/go-git v4.0.0-rc9.0.20190209005256-3a78daa8ce8e+incompatible
	mvdan.cc/xurls/v2 => github.com/keybase/xurls/v2 v2.0.1-0.20190725180013-1e015cacd06c
)
