package updater

import (
	"path/filepath"

	"github.com/mholt/archiver"
)

func Update(bundle, instDir, appName string) (error, string) {
	appExecName := appName + ".exe"
  appExec := filepath.Join(instDir, appExecName)

  err := archiver.Zip.Open(bundle, instDir)
	if err != nil {
		return err, appExec
	}

  return nil, appExec
}
