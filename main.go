package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"./archiver"
	"./open"
	"./xattr"
)

func main() {
	var bundle, instDir string

	flag.StringVar(&bundle, "bundle", "", "Path to the update package")
	flag.StringVar(&instDir, "inst-dir", "", "Path to the application install dir")
	flag.Parse()

	cwd, _ := os.Getwd()
	logfile, err := os.Create(filepath.Join(cwd, "updater.log"))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	logger := log.New(logfile, "", log.LstdFlags)

	appName := "my_app"

	var appExecName string
	switch runtime.GOOS {
	case "windows":
		appExecName = appName + ".exe"
	case "darwin":
		appExecName = appName + ".app"
	}

	appExec := filepath.Join(instDir, appExecName)

	var appDir string
	switch runtime.GOOS {
	case "windows":
		appDir = instDir
	case "darwin":
		appDir = appExec
	}

	tempDir, err := ioutil.TempDir("", "my-app-updates-")
	if err != nil {
		logger.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logger.Println("Update file: ", bundle)
	logger.Println("Application installed in: ", instDir)
	logger.Println("Application executable: ", appExec)
	logger.Println("Temporary directory: ", tempDir)

	logger.Println("Unpacking...")

	err = archiver.Unzip(bundle, tempDir)
	if err != nil {
		logger.Fatal(err)
	}

	if runtime.GOOS == "darwin" {
		appBak := appExec + ".bak"

		logger.Println("Creating application backup: ", appBak)

		err = os.Rename(appExec, appBak)
		if err != nil {
			logger.Fatal(err)
		}

		updateFiles := filepath.Join(tempDir, appExecName)

		logger.Println("Updating application from: ", updateFiles)

		err = os.Rename(updateFiles, appExec)
		if err != nil {
			logger.Println("Removing junk files...")
			os.RemoveAll(appExec)

			logger.Println("Restoring backup...")
			os.Rename(appBak, appExec)

			logger.Fatal(err)
		}

		logger.Println("Application updated!")
		logger.Println("Removing backup...")

		err = os.RemoveAll(appBak)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Removing update file...")

		err = os.RemoveAll(bundle)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Dropping extended attributes...")
		err = xattr.Remove(appExec, "com.apple.quarantine")
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Restarting application.")

		open.Start(appExec)
		os.Exit(0)
	}

	if runtime.GOOS == "windows" {
		appBak := appDir + ".bak"

		logger.Println("Creating application backup: ", appBak)

		err = os.Rename(appDir, appBak)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Updating application from: ", tempDir)

		err = os.Rename(tempDir, appDir)
		if err != nil {
			logger.Println("Removing junk files...")
			os.RemoveAll(appDir)

			logger.Println("Restoring backup...")
			os.Rename(appBak, appDir)

			logger.Fatal(err)
		}

		logger.Println("Application updated!")
		logger.Println("Removing backup...")

		err = os.RemoveAll(appBak)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Removing update file...")

		err = os.RemoveAll(bundle)
		if err != nil {
			logger.Fatal(err)
		}

		logger.Println("Restarting application.")

		open.Start(appExec)
		os.Exit(0)
	}

	logger.Fatal("Unknown OS")
	os.Exit(1)
}
