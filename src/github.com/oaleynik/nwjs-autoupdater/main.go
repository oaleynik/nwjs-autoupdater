package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
	"github.com/oaleynik/nwjs-autoupdater/updater"
)

func main() {
	var bundle, instDir string

	flag.StringVar(&bundle, "bundle", "", "Path to the update package")
	flag.StringVar(&instDir, "inst-dir", "", "Path to the application install dir")
	flag.Parse()

	appName := "my_app"

	cwd, _ := os.Getwd()
	logfile, err := os.Create(filepath.Join(cwd, "updater.log"))
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	logger := log.New(logfile, "", log.LstdFlags)

	var appExec string;
	err, appExec = updater.Update(bundle, instDir, appName)
	if err != nil {
		logger.Fatal(err)
	}

	open.Start(appExec)
}
