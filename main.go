// Copyright 2020 Noval Agung Prayogo. All rights reserved.
// Use of this source code is governed by a CC BY-NC-SA 4.0-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// this var used to store the shortcut file path,
// taken from argument during `go run`
var shortcutFilePath = ""

// these vars below will have values injected during build time
var isRuntime = ""
var osName = ""
var url = ""               // during .exe creation
var icon = ""              // during .exe creation
var bashScriptCommand = "" // during linux/unix/mac executable file creation

func main() {

	// detect wheter this app is running for run time or build time
	if isRuntime == "true" {
		runExecutable()
	} else {
		buildExecutable()
	}
}

// exec url/file
func runExecutable() {
	if osName == "windows" {
		execCommand("rundll32", "url.dll,FileProtocolHandler", url)
	} else {
		execCommand(bashScriptCommand)
	}
}

// build dot exe file
func buildExecutable() {
	if len(os.Args) < 2 {
		forceExitIfError("first argument should be a path of the shortcut file")
	}
	shortcutFilePath = os.Args[1]

	// construct the executable file namen
	filename := filepath.Base(shortcutFilePath)
	fileExt := strings.ToLower(filepath.Ext(filename))
	shouldBecomeDotExe := true
	if fileExt == ".lnk" || fileExt == ".url" || fileExt == ".cda" {
		filename = filename[0:len(filename)-len(fileExt)] + ".exe"
	} else if fileExt == ".desktop" {
		filename = filename[0 : len(filename)-len(fileExt)]
		shouldBecomeDotExe = false
	} else {
		forceExitIfError("unsupported file. file must be one of these: .lnk, .url, .cda, .desktop")
	}

	// get file metadata
	loadShortcutFileMetadata()

	// build the executable
	//   => .lnk, .url, .cda will be converted into .exe
	//   => .desktop will be converted into linux/unix/mac executable file
	if shouldBecomeDotExe {

		// fail if gopath is not set
		if os.Getenv("GOPATH") == "" {
			forceExitIfError("GOPATH env var is required to be set")
		}

		// fail if go binary is not callable
		if _, err := exec.LookPath("go"); err != nil {
			forceExitIfError("the go binary is required to be added into PATH env var")
		}

		// fail if git binary is not calleble, required by `go get` command
		if _, err := exec.LookPath("git"); err != nil {
			forceExitIfError("the git binary is required to be added into PATH env var")
		}

		iconFlags := ""
		if icon != "" {
			// go get the rsrc lib
			execCommand("go", "install", "github.com/akavel/rsrc")

			// create the rsrc.syso to set the icon of upcoming executable
			gopath := os.Getenv("GOPATH")
			rsrcFilename := "rsrc.syso"
			rsrcExecFile := filepath.Join(gopath, "bin", strings.Split(rsrcFilename, ".")[0])
			execCommand(rsrcExecFile, "-ico", icon)

			// remove syso file later
			defer os.Remove(rsrcFilename)

			// set ldflags
			iconFlags = fmt.Sprintf(` -X "main.icon=%s"`, icon)
		}

		// remove existing file
		os.Remove(filename)

		// build the .exe file
		execCommand(
			"go",
			"build",
			"-ldflags", fmt.Sprintf(`-X "main.isRuntime=true" -X "main.osName=%s" -X "main.url=%s"`, runtime.GOOS, url)+iconFlags,
			"-o", filename,
		)
	} else {

		// build the executable file
		execCommand(
			"go",
			"build",
			"-ldflags", fmt.Sprintf(`-X "main.isRuntime=true" -X "main.osName=%s" -X "main.bashScriptCommand=%s"`, runtime.GOOS, url, bashScriptCommand),
			"-o", filename,
		)
	}

	fmt.Printf("  => result: executable %s is successfully generated", filename)
}

// get metadata of shortcut file
func loadShortcutFileMetadata() {
	buf, err := ioutil.ReadFile(shortcutFilePath)
	forceExitIfError(err)

	lines := make(map[string]string)
	for _, o := range strings.Split(string(buf), "\n") {

		parts := strings.Split(strings.TrimSpace(o), "=")
		key := strings.ToLower(parts[0])
		if key == "iconfile" {
			key = "icon"
		}

		if len(parts) == 1 {
			lines[key] = ""
		} else {
			lines[key] = parts[1]
		}
	}

	url = lines["url"]
	icon = lines["icon"]
	bashScriptCommand = lines["exec"]
}

// utility func to force fail whenever there is error occuring
func forceExitIfError(message interface{}) {
	if message == nil {
		return
	}

	messageString := ""
	switch message.(type) {
	case string:
		messageString = message.(string)
	case error:
		messageString = message.(error).Error()
	}

	fmt.Println(messageString)
	os.Exit(0)
}

func execCommand(command ...string) {
	if runtime.GOOS == "windows" {
		command = append([]string{"cmd", "/C"}, command...)
	} else {
		command = append([]string{"/bin/sh", "-c"}, command...)
	}

	log.Println(strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		forceExitIfError(fmt.Sprintf("  => error: %s %s", err.Error(), stderr.String()))
	}
}
