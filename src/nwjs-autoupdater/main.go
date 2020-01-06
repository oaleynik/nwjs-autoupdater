package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
	"github.com/skratchdot/open-golang/open"
	"nwjs-autoupdater/updater"
)

func main() {
	var bundle, instDir, appName string

	flag.StringVar(&bundle, "bundle", "", "Path to the update package")
	flag.StringVar(&instDir, "inst-dir", "", "Path to the application install dir")
	flag.StringVar(&appName, "app-name", "my_app", "Application executable name")
	flag.Parse()

	cwd, _ := os.Getwd()
	logfile, err := os.Create(filepath.Join(cwd, "updater.log"))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	time.Sleep(2 * time.Second)

	logger := log.New(logfile, "", log.LstdFlags)

	var appExec string;
	err, appExec = updater.Update(bundle, instDir, appName)
	if err != nil {
		logger.Fatal(err)
	}

	open.Start(appExec)
}