// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package keybase

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/kardianos/osext"
	"github.com/keybase/go-updater"
	"github.com/keybase/go-updater/command"
	"github.com/keybase/go-updater/util"
	"golang.org/x/sys/windows"
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// FOLDERID_LocalAppData
// F1B32785-6FBA-4FCF-9D55-7B8E7F157091
var (
	folderIDLocalAppData = guid{0xF1B32785, 0x6FBA, 0x4FCF, [8]byte{0x9D, 0x55, 0x7B, 0x8E, 0x7F, 0x15, 0x70, 0x91}}
	folderIDSystem       = guid{0x1AC14E77, 0x02E7, 0x4E5D, [8]byte{0xB7, 0x44, 0x2E, 0xB1, 0xAE, 0x51, 0x98, 0xB7}}
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

func detectDokanDll(log Log) bool {
	dir, err := getDataDir(folderIDSystem)
	if err != nil {
		log.Infof("detectDokanDll error getting system directory: %v", err)
		return false
	}

	exists, _ := util.FileExists(filepath.Join(dir, "dokan1.dll"))
	log.Infof("detectDokanDll: returning %v", exists)
	return exists
}

type regUninstallGetter func(string, Log) bool

// CheckCanBeSilent - Takes an incoming installer's driver
// uninstall codes, which will be in the registry if the same
// version is already present.
// If it's not installed, check if dokan1.dll is present, which will indicate
// some other version is present.
func CheckCanBeSilent(dokanCodeX86 string, dokanCodeX64 string, log Log, regFunc regUninstallGetter) (bool, error) {
	codeFound := false
	var err error

	// Also look in the registry whether a reboot is pending from a previous installer.
	// Not finding the key is not an error.
	if rebootPending, err := checkRebootPending(false, log); rebootPending && err == nil {
		return false, nil
	}

	if rebootPending, err := checkRebootPending(true, log); rebootPending && err == nil {
		return false, nil
	}

	if regFunc(dokanCodeX86, log) || regFunc(dokanCodeX64, log) {
		codeFound = true
	} else {
		// If we don't find our dokan installed, see whether another one is,
		// and allow silent update if not.
		if !detectDokanDll(log) {
			codeFound = true
		}
	}

	log.Infof("CheckCanBeSilent: returning %v", codeFound)
	return codeFound, err
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

func (c context) Apply(update updater.Update, options updater.UpdateOptions, tmpDir string) error {
	if update.Asset == nil || update.Asset.LocalPath == "" {
		return fmt.Errorf("No asset")
	}
	var args []string
	auto, _ := c.config.GetUpdateAuto()
	if auto && !c.config.GetUpdateAutoOverride() {
		args = append(args, "/quiet", "/norestart")
	}
	_, err := command.Exec(update.Asset.LocalPath, args, time.Hour, c.log)
	return err
}

func (c context) AfterApply(update updater.Update) error {
	return nil
}
