package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gocd/gocd-trial-launcher/gocd"
	"github.com/gocd/gocd-trial-launcher/trap"
	"github.com/gocd/gocd-trial-launcher/utils"
)

var (
	dbgFlg = flag.Bool(`X`, false, `Enables debug output`)
	verFlg = flag.Bool(`version`, false, `Displays versions and exits`)
	ansFlg = flag.Bool(`ansitest`, false, `Displays ansi escape sequence tests`)
	rstFlg = flag.Bool(`reset`, false, `Resets the test drive data back to its initial state`)
)

func main() {
	flag.Parse()
	utils.EnableDebug = *dbgFlg

	if *verFlg {
		utils.Die(0, versionInfo())
	}

	if *ansFlg {
		utils.Die(0, AnsiTest)
	}

	trap.Trap(cleanup, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	if !utils.AllDirsExist(servPkgDir, agntPkgDir, javaHome) {
		utils.Die(1, "This GoCD test drive archive is missing 1 or more dependencies in the `packages` directory.\nPlease extract a clean copy from the downloaded archive and try again.")
	}

	utils.Debug(`Setting JAVA_HOME: %q`, javaHome)
	_ = os.Setenv(`JAVA_HOME`, javaHome)

	if err := java.Verify(); err != nil {
		utils.Err("Error executing java binary at [%s].\nIt might be incompatible with your OS.\n\n  Cause: %v\n", java.Executable(), err)
	}

	if utils.TryConnect(gocd.BindHost, gocd.HttpPort) {
		utils.Die(1, `Port %d must be free to run this test drive.`, gocd.HttpPort)
	}

	if *rstFlg {
		if utils.IsDir(dataDir) {
			utils.Out("Clobbering exisiting local data directory %q", dataDir)

			if err := os.RemoveAll(dataDir); err != nil {
				utils.Debug("Unable to remove directory %q; please check your permissions:\n  Cause: %v", dataDir, err)
			}
		}
	}

	if err := utils.Unzip(configZip, baseDir); err != nil {
		utils.Debug("Unable to apply configurations. Cause: %v", err)
	}

	if err := utils.MkdirP(serverWd, agentWd); err != nil {
		utils.Die(1, "Could not create a local data directory; please check your file permissions:\n  Cause: %v", err)
	}

	gocd.PrintLogo()

	var err error
	serverCmd, err = gocd.StartServer(java, serverWd, filepath.Join(servPkgDir, "go.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD server.\n  Cause: %v", err)
		cleanup()
	}

	utils.WaitUntilPortAttached(gocd.HttpPort, `Waiting for GoCD to bootstrap`)

	utils.Out("\n")
	utils.Out("Server log directory: %q", filepath.Join(serverWd, `logs`))
	utils.Out("Agent log directory:  %q", filepath.Join(agentWd, `logs`))
	utils.Out("All data written to:  %q", dataDir)

	utils.OpenUrlInBrowser(gocd.WebUrl)

	utils.Out("")
	utils.WaitUntilResponseSuccess(gocd.WebUrl, `Wating for the GoCD server to finish initializing`)
	utils.Out("\nThe GoCD Server has started")
	agentCmd, err = gocd.StartAgentBootstrapper(java, agentWd, filepath.Join(agntPkgDir, "agent-bootstrapper.jar"))

	if err != nil {
		utils.Err("Could not start the GoCD agent.\n  Cause: %v", err)
		cleanup()
	}

	utils.Out("\nPress Ctrl-C to exit")

	trap.WaitForInterrupt()
}

func versionInfo() string {
	return fmt.Sprintf(`run-gocd %s %s (%s)`, Version, Platform, GitCommit)
}
