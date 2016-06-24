#!/usr/bin/env bash

set -f -u -e

echo "Current directory: `pwd`"

gopath=${GOPATH:-"$HOME/gopath"}

# To install gometalinter
if [[ $TRAVIS_OS_NAME == 'linux' ]]; then
  go get -u github.com/alecthomas/gometalinter
  gometalinter --install --update
  echo "Running gometalinter"
  gometalinter --deadline=300s --vendor --cyclo-over=20 --dupl-threshold=100 ./... || exit 1
  result"$?"
  echo "Result: $result"
fi

repo="github.com/keybase/go-updater"
cd "$gopath/src/$repo"

godirs=`go list ./... | grep -v /vendor/`
for i in $godirs; do
  if [ "$i" = "$repo" ]; then
    echo "$repo..."
    go test -timeout 5m -coverprofile="main.coverprofile"
  else
    package=${i##*/}
    echo "$repo ($package)..."
    go test -timeout 5m -coverprofile="$package.coverprofile" ./"$package"
  fi
done

"$gopath/bin/gover"
"$gopath/bin/goveralls" -coverprofile=gover.coverprofile -service=travis-ci
