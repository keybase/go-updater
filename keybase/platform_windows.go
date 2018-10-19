// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/process"
	"github.com/keybase/go-updater/util"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	// FOLDERID_LocalAppData
	// F1B32785-6FBA-4FCF-9D55-7B8E7F157091
	folderIDLocalAppData = guid{0xF1B32785, 0x6FBA, 0x4FCF, [8]byte{0x9D, 0x55, 0x7B, 0x8E, 0x7F, 0x15, 0x70, 0x91}}

	// FOLDERID_RoamingAppData
	// {3EB685DB-65F9-4CF6-A03A-E3EF65729F3D}
	folderIDRoamingAppData = guid{0x3EB685DB, 0x65F9, 0x4CF6, [8]byte{0xA0, 0x3A, 0xE3, 0xEF, 0x65, 0x72, 0x9F, 0x3D}}
)

var (
	modShell32               = windows.NewLazySystemDLL("Shell32.dll")
	modOle32                 = windows.NewLazySystemDLL("Ole32.dll")
	procSHGetKnownFolderPath = modShell32.NewProc("SHGetKnownFolderPath")
	procCoTaskMemFree        = modOle32.NewProc("CoTaskMemFree")
)

func coTaskMemFree(pv uintptr) {
	syscall.Syscall(procCoTaskMemFree.Addr(), 1, uintptr(pv), 0, 0)
	return
}

func getDataDir(id guid) (string, error) {

	var pszPath uintptr
	r0, _, _ := procSHGetKnownFolderPath.Call(uintptr(unsafe.Pointer(&id)), uintptr(0), uintptr(0), uintptr(unsafe.Pointer(&pszPath)))
	if r0 != 0 {
		return "", errors.New("can't get known folder")
	}

	defer coTaskMemFree(pszPath)

	// go vet: "possible misuse of unsafe.Pointer"
	folder := syscall.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(pszPath))[:])

	if len(folder) == 0 {
		return "", errors.New("can't get known folder")
	}

	return folder, nil
}

func localDataDir() (string, error) {
	return getDataDir(folderIDLocalAppData)
}

func roamingDataDir() (string, error) {
	return getDataDir(folderIDRoamingAppData)
}

func (c config) destinationPath() string {
	pathName, err := osext.Executable()
	if err != nil {
		c.log.Warningf("Error trying to determine our executable path: %s", err)
		return ""
	}
	dir, _ := filepath.Split(pathName)
	return dir
}

// Dir returns where to store config and log files
func Dir(appName string) (string, error) {
	dir, err := localDataDir()
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("No LocalDataDir")
	}
	if appName == "" {
		return "", fmt.Errorf("No app name for dir")
	}
	return filepath.Join(dir, appName), nil
}

// CacheDir returns where to store temporary files
func CacheDir(appName string) (string, error) {
	return Dir(appName)
}

// LogDir is where to log
func LogDir(appName string) (string, error) {
	return Dir(appName)
}

func (c config) osVersion() string {
	result, err := command.Exec("cmd", []string{"/c", "ver"}, 5*time.Second, c.log)
	if err != nil {
		c.log.Warningf("Error trying to determine OS version: %s (%s)", err, result.CombinedOutput())
		return ""
	}
	return strings.TrimSpace(result.Stdout.String())
}

func (c config) osArch() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `Hardware\Description\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return err.Error()
	}
	defer k.Close()

	s, _, err := k.GetStringValue("Identifier")
	if err != nil {
		return err.Error()
	}
	words := strings.Fields(s)
	if len(words) < 1 {
		return "empty"
	}
	return words[0]
}

func (c config) notifyProgram() string {
	// No notify program for Windows
	return runtime.GOARCH
}

func (c *context) BeforeUpdatePrompt(update updater.Update, options updater.UpdateOptions) error {
	return nil
}

func (c config) promptProgram() (command.Program, error) {
	destinationPath := c.destinationPath()
	if destinationPath == "" {
		return command.Program{}, fmt.Errorf("No destination path")
	}

	return command.Program{
		Path: filepath.Join(destinationPath, "prompter.exe"),
	}, nil
}

func (c context) UpdatePrompt(update updater.Update, options updater.UpdateOptions, promptOptions updater.UpdatePromptOptions) (*updater.UpdatePromptResponse, error) {
	promptProgram, err := c.config.promptProgram()
	if err != nil {
		return nil, err
	}

	if promptOptions.OutPath == "" {
		promptOptions.OutPath, err = util.WriteTempFile("updatePrompt", []byte{}, 0700)
		if err != nil {
			return nil, err
		}
		defer util.RemoveFileAtPath(promptOptions.OutPath)
	}

	promptJSONInput, err := c.promptInput(update, options, promptOptions)
	if err != nil {
		return nil, fmt.Errorf("Error generating input: %s", err)
	}

	_, err = command.Exec(promptProgram.Path, promptProgram.ArgsWith([]string{string(promptJSONInput)}), time.Hour, c.log)
	if err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	result, err := c.updaterPromptResultFromFile(promptOptions.OutPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading result: %s", err)
	}
	return c.responseForResult(*result)
}

// updaterPromptResultFromFile gets the result from path decodes it
func (c context) updaterPromptResultFromFile(path string) (*updaterPromptInputResult, error) {
	resultRaw, err := util.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result updaterPromptInputResult
	if err := json.Unmarshal(resultRaw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c context) PausedPrompt() bool {
	return false
}

type componentProductFunc func(componentKey registry.Key, productValueName, componentPath string)

type ComponentsChecker struct {
	context
	RegAccess    uint32
	RegWow       uint32
	PerComponent componentProductFunc
}

func (i *ComponentsChecker) deleteProductsFunc(componentKey registry.Key, productValueName, componentPath string) {
	i.log.Infof("Found Keybase component %s, deleting\n", componentPath)
	err := componentKey.DeleteValue(productValueName)
	if err != nil {
		i.log.Infof("Error DeleteValue %s: %s\n", productValueName, err.Error())
	}
}

// checkRegistryComponents returns true if any component has more than one keybase product code
func (c *ComponentsChecker) checkRegistryComponents() (result bool) {
	// e.g.
	// [HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Installer\UserData\S-1-5-21-2398092721-582601651-115936829-1001\Components\024E69EF1A837C752BFB37F494D86925]
	// "D6A082CFDEED2984C8688664C76174BC"="C:\\Users\\chris\\AppData\\Local\\Keybase\\Gui\\resources\\app\\images\\icons\\icon-facebook-visibility.gif"
	// "50DC76D18793BC24DA7D4D681AE74262"="C:\\Users\\chris\\AppData\\Local\\Keybase\\Gui\\resources\\app\\images\\icons\\icon-facebook-visibility.gif"

	readAccess := registry.ENUMERATE_SUB_KEYS | registry.QUERY_VALUE | c.RegWow

	rootName := "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Installer\\UserData"

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, rootName, readAccess)
	if err != nil {
		c.log.Infof("Error opening uninstall subkeys: %s\n", err.Error())
		return
	}
	defer k.Close()

	UIDs, err := k.ReadSubKeyNames(-1)
	if err != nil {
		c.log.Infof("Error reading subkeys: %s\n", err.Error())
		return
	}
	for _, UID := range UIDs {
		componentsKey, err := registry.OpenKey(k, UID+"\\Components", readAccess)
		if err != nil {
			c.log.Infof("Error opening subkey %s: %s\n", UID+"\\Components", err.Error())
			continue
		}

		componentKeyNames, err := componentsKey.ReadSubKeyNames(-1)
		if err != nil {
			c.log.Infof("Error reading subkeys: %s\n", err.Error())
			continue
		}

		for _, componentKeyName := range componentKeyNames {
			componentKey, err := registry.OpenKey(componentsKey, componentKeyName, readAccess|c.RegAccess)
			if err != nil {
				c.log.Infof("Error opening subkey %s: %s\n", componentKeyName, err.Error())
				// No need to list all the components we couldn't open in write mode.
				// This is expected when run without elevated permissions.
				c.log.Infof("skipping subsequent subkeys of  %s\n", UID+"\\Components")
				continue
			}

			productValueNames, err := componentKey.ReadValueNames(-1)
			if err != nil {
				c.log.Infof("Error reading values: %s\n", err.Error())
				continue
			}

			for n, productValueName := range productValueNames {
				componentPath, _, err := componentKey.GetStringValue(productValueName)
				if err == nil && strings.Contains(componentPath, "\\AppData\\Local\\Keybase\\") {
					if c.PerComponent != nil {
						c.PerComponent(componentKey, productValueName, componentPath)
					}
					if n > 0 {
						result = true
						c.log.Infof("Found multiple Keybase product codes on %s\n", componentPath)
					}
				}
			}
			componentKey.Close()
		}
		componentsKey.Close()
	}
	return result
}

type KeybaseCommand string

const (
	KeybaseCommandStart KeybaseCommand = "watchdog2"
	KeybaseCommandStop  KeybaseCommand = "stop"
)

func (c context) runKeybase(cmd KeybaseCommand) {
	path, err := Dir("Keybase")
	if err != nil {
		c.log.Infof("Error getting Keybase directory: %s", err.Error())
		return
	}

	args := []string{filepath.Join(path, "keybase.exe"), "ctl", string(cmd)}

	_, err = command.Exec(filepath.Join(path, "keybaserq.exe"), args, time.Minute, c.log)
	if err != nil {
		c.log.Infof("Error %s'ing keybase", cmd, err.Error())
	}
}

func (c context) deleteProductFiles() {
	path, err := Dir("Keybase")
	if err != nil {
		c.log.Infof("Error getting Keybase directory: %s", err.Error())
		return
	}
	c.stopKeybaseProcesses()

	err = os.RemoveAll(filepath.Join(path, "Gui"))
	if err != nil {
		c.log.Infof("Error removing Gui directory: %s", err.Error())
	}

	files, err := filepath.Glob(filepath.Join(path, "*.exe"))
	if err != nil {
		c.log.Infof("Error getting exe files: %s", err.Error())
	} else {
		for _, f := range files {
			c.log.Infof("Removing %s", f)
			if err = os.Remove(f); err != nil {
				c.log.Infof("Error removing file: %s", err.Error())
			}
		}
	}
}

// DeepClean is only invoked from the command line, for now.
// Eventually we may need to do full uninstalls but that is kind of risky
func (c context) DeepClean() {
	i := &ComponentsChecker{context: c, RegAccess: registry.SET_VALUE}
	i.PerComponent = i.deleteProductsFunc
	i.checkRegistryComponents()
	i.RegWow = registry.WOW64_32KEY
	i.checkRegistryComponents()
	c.deleteProductFiles()
}

func (c context) Apply(update updater.Update, options updater.UpdateOptions, tmpDir string) error {
	skipSilent := false
	if update.Asset == nil || update.Asset.LocalPath == "" {
		return fmt.Errorf("No asset")
	}
	c.stopKeybaseProcesses()
	if c.config.GetLastAppliedVersion() == update.Version {
		c.log.Info("Previously applied version detected")
		c.config.SetLastAppliedVersion("")
		skipSilent = true
	}

	runCommand := update.Asset.LocalPath
	args := []string{}
	if strings.HasSuffix(runCommand, "msi") || strings.HasSuffix(runCommand, "MSI") {
		args = append([]string{
			"/i",
			runCommand,
			"/log",
			filepath.Join(
				os.TempDir(),
				fmt.Sprintf("KeybaseMsi_%d%02d%02d%02d%02d%02d.log",
					time.Now().Year(),
					time.Now().Month(),
					time.Now().Day(),
					time.Now().Hour(),
					time.Now().Minute(),
					time.Now().Second(),
				),
			),
		}, args...)
		runCommand = "msiexec.exe"
	}
	auto, _ := c.config.GetUpdateAuto()
	if auto && !c.config.GetUpdateAutoOverride() && !skipSilent {
		args = append(args, "/quiet", "/norestart")
	}
	c.config.SetLastAppliedVersion(update.Version)
	_, err := command.Exec(runCommand, args, time.Hour, c.log)
	return err
}

// Note that when a Windows installer runs, it kills the running updater, even
// before AfterApply() runs
func (c context) AfterApply(update updater.Update) error {
	return nil
}

// app-state.json is written in the roaming settings directory, which
// seems to be where Electron chooses
func (c context) GetAppStatePath() string {
	roamingDir, _ := roamingDataDir()
	return filepath.Join(roamingDir, "Keybase", "app-state.json")
}

func (c context) IsCheckCommand() bool {
	return c.isCheckCommand
}

// copied from watchdog
func (c context) stopKeybaseProcesses() error {
	path, err := Dir("Keybase")
	if err != nil {
		c.log.Infof("Error getting Keybase directory: %s", err.Error())
		return err
	}

	c.runKeybase(KeybaseCommandStop)
	time.Sleep(time.Second)

	// Terminate any executing processes
	ospid := os.Getpid()

	exes, err := filepath.Glob(filepath.Join(path, "*.exe"))
	if err != nil {
		c.log.Errorf("Unable to glob exe files: %s", err)
	}
	guiExes, err := filepath.Glob(filepath.Join(path, "Gui", "*.exe"))
	if err != nil {
		c.log.Errorf("Unable to glob exe files: %s", err)
	} else {
		exes = append(exes, guiExes...)
	}

	c.log.Infof("Terminating any existing programs we will be updating")
	for _, program := range exes {
		matcher := process.NewMatcher(program, process.PathEqual, c.log)
		matcher.ExceptPID(ospid)
		c.log.Infof("Terminating %s", program)
		process.TerminateAll(matcher, time.Second, c.log)
	}
	return nil
}
