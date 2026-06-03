package cmd

import (
	"os"

	"github.com/sevlyar/go-daemon"
)

var (
	programName = "nexus-cli"
	pidFile     = ".pid"
	logFile     = ".log"
	cntxt       = &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
	}
)

func isRunning() (*os.Process, bool) {
	process, err := cntxt.Search()
	if err != nil || process == nil {
		return nil, false
	}
	return process, true
}
