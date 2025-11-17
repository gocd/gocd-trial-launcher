package main

import (
	"os/exec"
	"path/filepath"

	"github.com/gocd/gocd-trial-launcher/utils"
)

var (
	baseDir    = utils.BaseDir()
	packageDir = filepath.Join(baseDir, `packages`)
	dataDir    = filepath.Join(baseDir, `data`)
	servPkgDir = filepath.Join(packageDir, `go-server`)
	agntPkgDir = filepath.Join(packageDir, `go-agent`)

	configZip = filepath.Join(packageDir, `cfg.zip`)

	javaHome = filepath.Join(packageDir, `jre`)
	java     = utils.NewJava(javaHome)

	serverWd = filepath.Join(dataDir, `server`)
	agentWd  = filepath.Join(dataDir, `agent`)
)

// These should be set by the linker at build time
var (
	Version   = `devbuild`
	GitCommit = `unknown`
	Platform  = `devbuild`
)

var agentCmd *exec.Cmd
var serverCmd *exec.Cmd
