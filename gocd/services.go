package gocd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocd/gocd-trial-launcher/utils"
)

const (
	HttpPort = 8153
	BindHost = `localhost`
)

//goland:noinspection HttpUrlsUsage
var (
	WebUrl           = `http://` + BindHost + `:` + strconv.Itoa(HttpPort)
	AgentRegisterUrl = `http://` + BindHost + `:` + strconv.Itoa(HttpPort) + `/go`
)

func StartServer(java *utils.Java, workDir, jar string) (*exec.Cmd, error) {
	configDir := filepath.Join(workDir, "config")
	configFile := filepath.Join(configDir, "cruise-config.xml")
	tmpDir := filepath.Join(workDir, "tmp")
	logDir := filepath.Join(workDir, "logs")
	logFile := filepath.Join(logDir, "stdout.log")

	if err := utils.MkdirP(configDir, tmpDir, logDir); err != nil {
		return nil, err
	}

	props := utils.JavaProps{
		"cruise.config.dir":            configDir,
		"cruise.config.file":           configFile,
		"java.io.tmpdir":               tmpDir,
		"gocd.redirect.stdout.to.file": logFile,
	}

	if err := mergeExtraProperties(props, jar); err != nil {
		return nil, err
	}

	return startJavaApp(java, "server", workDir, props, nil,
		"-Xmx1024m",
		"--add-opens=java.base/java.lang=ALL-UNNAMED", // Match https://github.com/gocd/gocd/blob/776fc1d4b8585c489c894a68aac22b4bfa2550e9/buildSrc/src/main/groovy/com/thoughtworks/go/build/InstallerTypeServer.groovy#L45-L54
		"--add-opens=java.base/java.util=ALL-UNNAMED",
		"--enable-native-access=ALL-UNNAMED",
		"--sun-misc-unsafe-memory-access=allow",
		"-XX:+IgnoreUnrecognizedVMOptions",
		"-jar", jar,
		"-server")
}

func StartAgentBootstrapper(java *utils.Java, workDir, jar string) (*exec.Cmd, error) {
	tmpDir := filepath.Join(workDir, "tmp")
	logDir := filepath.Join(workDir, "logs")
	logFile := filepath.Join(logDir, "stdout.log")

	if err := utils.MkdirP(tmpDir, logDir); err != nil {
		return nil, err
	}

	props := utils.JavaProps{
		"java.io.tmpdir":               tmpDir,
		"gocd.redirect.stdout.to.file": logFile,
		"gocd.agent.log.dir":           logDir,
	}

	if err := mergeExtraProperties(props, jar); err != nil {
		return nil, err
	}

	agentStartupArgsEnv := utils.EnvVars{ // Match https://github.com/gocd/gocd/blob/bcf11e1f6e7f5f6bef7e875f957fac0172277a6e/buildSrc/src/main/groovy/com/thoughtworks/go/build/InstallerTypeAgent.groovy#L55-L63
		"AGENT_STARTUP_ARGS": strings.Join([]string{
			"--enable-native-access=ALL-UNNAMED",
			"--sun-misc-unsafe-memory-access=allow",
			"-XX:+IgnoreUnrecognizedVMOptions",
		}, " "),
	}

	return startJavaApp(java, "agent", workDir, props, agentStartupArgsEnv, "-Xmx256m", "-jar", jar, "-serverUrl", AgentRegisterUrl)
}

func StopServer(cmd *exec.Cmd) {
	if cmd != nil {
		pidFile := filepath.Join(cmd.Dir, "server.pid")

		stopApp(cmd, pidFile, "server")
	}
}

func StopAgent(cmd *exec.Cmd) {
	if cmd != nil {
		pidFile := filepath.Join(cmd.Dir, "agent.pid")

		stopApp(cmd, pidFile, "agent")
	}
}

func mergeExtraProperties(props utils.JavaProps, jar string) error {
	utils.Debug(`Checking for extra properties for app at %q`, jar)
	path := filepath.Join(filepath.Dir(jar), `extra-props.yaml`)

	if utils.IsFile(path) {
		utils.Debug(`Reading java properties from %q`, path)

		if extras, err := utils.PropsFromYaml(path); err == nil {
			for k, v := range extras {
				utils.Debug(`Found property %s=%s`, k, v)
				props[k] = v
			}
		} else {
			utils.Debug(`Error while extracting properties from %q: %v`, path, err)
			return err
		}
	}
	return nil
}

func startJavaApp(java *utils.Java, serviceName string, workDir string, properties utils.JavaProps, additionalEnv utils.EnvVars, args ...string) (*exec.Cmd, error) {
	cmd := java.Build(properties, additionalEnv, args...)

	utils.EnablePgid(cmd)

	cmd.Dir = workDir
	pidFile := filepath.Join(workDir, serviceName+".pid")

	utils.Out("\nStarting the GoCD %s...", serviceName)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	utils.Debug(`%s PID: %d, writing to pidfile: %q`, serviceName, cmd.Process.Pid, pidFile)

	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		return nil, err
	}

	return cmd, nil
}

func stopApp(cmd *exec.Cmd, pidFile, serviceName string) {
	utils.Debug(`Ending %s process %d`, serviceName, cmd.Process.Pid)

	if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
		utils.Out("Stopping GoCD %s...", serviceName)

		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			if err = cmd.Process.Kill(); err != nil {
				utils.Err("Unable to stop the GoCD test drive. See PID: %d", cmd.Process.Pid)
			}
		}
	}

	err := utils.KillPgid(cmd)
	if err != nil {
		utils.Err(`Could not kill %s process %d; continuing anyway...`, serviceName, cmd.Process.Pid)
	}

	if pidFile != "" && utils.IsExist(pidFile) {
		utils.Debug(`Removing pidfile: %q`, pidFile)

		if err := os.Remove(pidFile); err != nil {
			utils.Err("Failed to remove pidfile %s.\n  Cause: %v", pidFile, err)
		}
	}
}
