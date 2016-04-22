#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

app_name="Updater"
plist="$dir/Updater/Info.plist"
scheme="Updater"

build_dir="$dir/build"
archive_path="$build_dir/$app_name.xcarchive"

app_path="/Applications/Keybase.app/Contents/Resources/$app_name.app"

xcode_configuration="Release"
code_sign_identity="Developer ID Application: Keybase, Inc. (99229SGT5K)"

mkdir -p "$build_dir"

echo "Plist: $plist"
app_version="`/usr/libexec/plistBuddy -c "Print :CFBundleShortVersionString" $plist`"

echo "Archiving"
xcodebuild archive -scheme "$scheme" -project "$dir/Updater.xcodeproj" -configuration "$xcode_configuration" -archivePath "$archive_path" | xcpretty -c

echo "Exporting"
rm -rf "$app_path"
xcodebuild -exportArchive -archivePath "$archive_path" -exportFormat app -exportPath "$app_path" | xcpretty -c
echo "Exported to $app_path"
