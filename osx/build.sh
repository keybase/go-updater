#!/bin/bash

set -e -u -o pipefail # Fail on error

dir=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$dir"

app_name="KeybaseUpdater"
plist="$dir/Updater/Info.plist"
scheme="Updater"
code_sign_identity=${CODE_SIGN_IDENTITY:-"Developer ID Application: Keybase, Inc. (99229SGT5K)"}
xcode_configuration="Release"
install_app_path="/Applications/Keybase.app/Contents/Resources/$app_name.app"

build_dir="$dir/build"
mkdir -p "$build_dir"
archive_path="$build_dir/$app_name.xcarchive"

echo "Plist: $plist"
app_version="`/usr/libexec/plistBuddy -c "Print :CFBundleShortVersionString" $plist`"

echo "Archiving"
xcodebuild archive -scheme "$scheme" -project "$dir/Updater.xcodeproj" -configuration "$xcode_configuration" -archivePath "$archive_path" | xcpretty -c

echo "Exporting"
tmp_dir="/tmp"
tmp_app_path="$tmp_dir/$app_name.app"
rm -rf "$tmp_app_path"
xcodebuild -exportArchive -archivePath "$archive_path" -exportFormat app -exportPath "$tmp_app_path" | xcpretty -c
echo "Exported to $tmp_app_path"

echo "Codesigning with $code_sign_identity"
codesign --verbose --force --deep --sign "$code_sign_identity" "$tmp_app_path"
echo "Checking codesigning..."
codesign -dvvvv "$tmp_app_path"
echo " "
spctl --assess --verbose=4 "$tmp_app_path"
echo " "

cd "$tmp_dir"
tgz="$app_name-$app_version-darwin.tgz"
echo "Packing $tgz"
tar zcvpf "$tgz" "$app_name.app"
echo "Created $tmp_dir/$tgz"

rm -rf "$install_app_path"
cp -R "$tmp_app_path" "$install_app_path"
echo "Copied $tmp_app_path to $install_app_path"
