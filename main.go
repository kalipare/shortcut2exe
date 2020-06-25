// Copyright 2020 Noval Agung Prayogo. All rights reserved.
// Use of this source code is governed by a CC BY-NC-SA 4.0-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// this var used to store the shortcut file path, taken from argument during `go run`
var shortcutFilePath = ""

// these vars below will have values injected during build time, for .exe creation
var isRuntime = ""
var url = ""
var icon = ""

// this var below will be used to store the bash script command, for .sh creation
var bashScriptCommand = ""

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
	err := exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	forceExitIfError(err)
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
		filename = filename[0:len(filename)-len(fileExt)] + ".sh"
		shouldBecomeDotExe = false
	} else {
		forceExitIfError("unsupported file. file must be one of these: .lnk, .url, .cda, .desktop")
	}

	// get file metadata
	loadShortcutFileMetadata()

	// build the executable
	//   => .lnk, .url, .cda will be converted into .exe
	//   => .desktop will be converted into .sh
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

		// go get the rsrc lib
		execCommand([]string{"go", "get", "github.com/akavel/rsrc"})

		// create the rsrc.syso to set the icon of upcoming executable
		gopath := os.Getenv("GOPATH")
		rsrcFilename := "rsrc.syso"
		rsrcExecFile := filepath.Join(gopath, "bin", strings.Split(rsrcFilename, ".")[0])
		execCommand([]string{rsrcExecFile, "-ico", icon})

		// remove existing file
		os.Remove(filename)

		// build the .exe file
		execCommand([]string{
			"go",
			"build",
			"-ldflags", fmt.Sprintf(`-X "main.isRuntime=true" -X "main.url=%s" -X "main.icon=%s"`, url, icon),
			"-o", filename,
		})

		// remove syso file
		os.Remove(rsrcFilename)
	} else {
		content := []byte(fmt.Sprintf("#!/bin/sh\n\n%s", bashScriptCommand))
		err := ioutil.WriteFile(filename, content, os.ModePerm)
		forceExitIfError(err)
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

func execCommand(command []string) {
	if runtime.GOOS == "windows" {
		command = append([]string{"cmd", "/C"}, command...)
	} else {
		command = append([]string{"/bin/sh", "-c"}, command...)
	}

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
