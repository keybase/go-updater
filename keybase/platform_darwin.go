// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/process"
	"github.com/keybase/go-updater/util"
)

// execPath returns the app bundle path where this executable is located
func (c config) execPath() string {
	path, err := osext.Executable()
	if err != nil {
		c.log.Warningf("Error trying to determine our executable path: %s", err)
		return ""
	}
	return path
}

// destinationPath returns the app bundle path where this executable is located
func (c config) destinationPath() string {
	return appBundleForPath(c.execPath())
}

func appBundleForPath(path string) string {
	if path == "" {
		return ""
	}
	paths := strings.SplitN(path, ".app", 2)
	// If no match, return ""
	if len(paths) <= 1 {
		return ""
	}
	return paths[0] + ".app"
}

func libraryDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, "Library"), nil
}

// Dir returns where to store config
func Dir(appName string) (string, error) {
	if appName == "" {
		return "", fmt.Errorf("No app name for dir")
	}
	libDir, err := libraryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(libDir, "Application Support", appName), nil
}

// LogDir is where to log
func LogDir(appName string) (string, error) {
	libDir, err := libraryDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(libDir, "Logs"), nil
}

func (c config) osVersion() string {
	result, err := command.Exec("/usr/bin/sw_vers", []string{"-productVersion"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, result.CombinedOutput())
		return ""
	}
	return strings.TrimSpace(result.Stdout.String())
}

func (c config) promptProgram() (command.Program, error) {
	destinationPath := c.destinationPath()
	if destinationPath == "" {
		return command.Program{}, fmt.Errorf("No destination path")
	}
	return command.Program{
		Path: filepath.Join(destinationPath, "Contents", "Resources", "KeybaseUpdater.app", "Contents", "MacOS", "Updater"),
	}, nil
}

func (c config) notifyProgram() string {
	// No notify program for Darwin
	return ""
}

func (c context) BeforeUpdatePrompt(update updater.Update, options updater.UpdateOptions) error {
	return nil
}

// UpdatePrompt is called when the user needs to accept an update
func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptProgram, err := c.config.promptProgram()
	if err != nil {
		return nil, err
	}
	return c.updatePrompt(promptProgram, update, options, promptOptions, time.Hour)
}

// PausedPrompt is called when the we can't update cause the app is in use.
// We return true if the use wants to cancel the update.
func (c context) PausedPrompt() bool {
	promptProgram, err := c.config.promptProgram()
	if err != nil {
		c.log.Warningf("Error trying to get prompt path: %s", err)
		return false
	}
	cancelUpdate, err := c.pausedPrompt(promptProgram, 5*time.Minute)
	if err != nil {
		c.log.Warningf("Error in paused prompt: %s", err)
		return false
	}
	return cancelUpdate
}

const serviceInBundlePath = "/Contents/SharedSupport/bin/keybase"
const kbfsInBundlePath = "/Contents/SharedSupport/bin/kbfs"

// Restart will stop the services and app, and then start the app.
// The supervisor/watchdog process is in charge of restarting the services.
func (c context) Restart() error {
	return c.restart(10*time.Second, time.Second)
}

// restart will stop the services and app, and then start the app.
// The supervisor/watchdog process is in charge of restarting the services.
// The wait is how log to wait for processes and the app to start before
// reporting that an error occurred.
func (c context) restart(wait time.Duration, delay time.Duration) error {
	appPath := c.config.destinationPath() // "/Applications/Keybase.app"
	if appPath == "" {
		return fmt.Errorf("No destination path for restart")
	}

	appBundleName := filepath.Base(appPath) // "Keybase.app"

	serviceProcPath := appBundleName + serviceInBundlePath
	kbfsProcPath := appBundleName + kbfsInBundlePath
	appProcPath := appBundleName + "/Contents/MacOS/"

	// Request the app to exit (instead of sending a SIGTERM)
	if appExitResult, err := command.Exec(c.config.keybasePath(), []string{"ctl", "app-exit"}, 5*time.Second, c.log); err != nil {
		c.log.Warningf("Error requesting app exit: %s (%s)", err, appExitResult.CombinedOutput())
	}
	// Give the app time to exit
	time.Sleep(time.Second)
	// SIGKILL the app if it failed to exit
	c.log.Infof("Killing (if failed to exit) %s", appProcPath)
	process.KillAll(process.NewMatcher(appProcPath, process.PathContains, c.log), c.log)

	c.log.Infof("Terminating %s", serviceProcPath)
	process.TerminateAll(process.NewMatcher(serviceProcPath, process.PathContains, c.log), time.Second, c.log)
	c.log.Infof("Terminating %s", kbfsProcPath)
	process.TerminateAll(process.NewMatcher(kbfsProcPath, process.PathContains, c.log), time.Second, c.log)

	if err := OpenAppDarwin(appPath, c.log); err != nil {
		c.log.Warningf("Error opening app: %s", err)
	}

	// Check to make sure processes restarted
	serviceProcErr := c.checkProcess(serviceProcPath, wait, delay)
	kbfsProcErr := c.checkProcess(kbfsProcPath, wait, delay)
	appProcErr := c.checkProcess(appProcPath, wait, delay)

	return util.CombineErrors(serviceProcErr, kbfsProcErr, appProcErr)
}

func (c context) checkProcess(match string, wait time.Duration, delay time.Duration) error {
	matcher := process.NewMatcher(match, process.PathContains, c.log)
	procs, err := process.FindProcesses(matcher, wait, delay, c.log)
	if err != nil {
		return fmt.Errorf("Error checking on process (%s): %s", match, err)
	}
	if len(procs) == 0 {
		return fmt.Errorf("No process found for %s", match)
	}
	return nil
}

// OpenAppDarwin starts an app
func OpenAppDarwin(appPath string, log process.Log) error {
	return openAppDarwin("/usr/bin/open", appPath, time.Second, log)
}

func openAppDarwin(bin string, appPath string, retryDelay time.Duration, log process.Log) error {
	tryOpen := func() error {
		result, err := command.Exec(bin, []string{appPath}, time.Minute, log)
		if err != nil {
			return fmt.Errorf("Open error: %s; %s", err, result.CombinedOutput())
		}
		return nil
	}
	// We need to try 10 times because Gatekeeper has some issues, for example,
	// http://www.openradar.me/23614087
	var err error
	for i := 0; i < 10; i++ {
		err = tryOpen()
		if err == nil {
			break
		}
		log.Errorf("Open error (trying again in %s): %s", retryDelay, err)
		time.Sleep(retryDelay)
	}
	return err
}

func (c context) check(sourcePath string, destinationPath string) error {
	// Check to make sure the update source path is a real directory
	ok, err := util.IsDirReal(sourcePath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Source path isn't a directory")
	}
	return nil
}

func (c context) Apply(update updater.Update, options updater.UpdateOptions, tmpDir string) error {
	if update.Asset == nil {
		return fmt.Errorf("No asset")
	}
	localPath := update.Asset.LocalPath
	destinationPath := options.DestinationPath

	// The file name we unzip over should match the (base) file in the destination path
	filename := filepath.Base(destinationPath)
	if err := util.UnzipOver(localPath, filename, destinationPath, c.check, tmpDir, c.log); err != nil {
		return err
	}

	// Update spotlight
	c.log.Debugf("Updating spotlight: %s", destinationPath)
	spotlightResult, spotLightErr := command.Exec("/usr/bin/mdimport", []string{destinationPath}, 20*time.Second, c.log)
	if spotLightErr != nil {
		c.log.Warningf("Error trying to update spotlight: %s, (%s)", spotLightErr, spotlightResult.CombinedOutput())
	}

	// Remove quantantine (if any)
	c.log.Debugf("Remove quarantine: %s", destinationPath)
	quarantineResult, quarantineErr := command.Exec("/usr/bin/xattr", []string{"-d", "com.apple.quarantine", destinationPath}, 20*time.Second, c.log)
	if quarantineErr != nil {
		c.log.Warningf("Error trying to remove quarantine: %s, (%s)", quarantineErr, quarantineResult.CombinedOutput())
	}

	return nil
}
