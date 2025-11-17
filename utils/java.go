package utils

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v2"
)

var EnableDebug = false

type JavaProps map[string]string
type EnvVars map[string]string

func (j *JavaProps) Args() []string {
	args := make([]string, 0)
	for k, v := range *j {
		args = append(args, `-D`+k+`=`+v)
	}
	return args
}

// Java Builds a `java` command invocation with a specific JAVA_HOME
type Java struct {
	Home       string
	executable string
}

func PropsFromYaml(yamlFile string) (JavaProps, error) {
	if file, err := os.Open(yamlFile); err == nil {
		dcr := yaml.NewDecoder(file)
		dcr.SetStrict(true)
		m := make(map[string]string)
		if err = dcr.Decode(&m); err == nil {
			return m, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (j *Java) Build(properties JavaProps, additionalEnv EnvVars, args ...string) *exec.Cmd {
	if len(properties) > 0 {
		args = append(properties.Args(), args...)
	}

	cmd := exec.Command(j.executable, args...)
	cmd.Env = os.Environ() // inherit env

	if len(additionalEnv) > 0 {
		for k, v := range additionalEnv {
			cmd.Env = append(cmd.Env, k+"="+v)
			Debug(`Using %s: %q`, k, v)
		}
	}

	if os.Getenv(`JAVA_HOME`) != j.Home { // ensure JAVA_HOME uses the configured Home
		cmd.Env = append(cmd.Env, `JAVA_HOME=`+j.Home)
	}

	Debug(`Using JAVA_HOME: %q`, j.Home)
	Debug("%s %v", j.executable, args)

	return cmd
}

func (j *Java) Verify() error {
	cmd := j.Build(nil, nil, `-version`)

	if EnableDebug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	return cmd.Run()
}

func (j *Java) Executable() string {
	return j.executable
}

func NewJava(javaHome string) *Java {
	java := filepath.Join(javaHome, `bin`, `java`)

	if `windows` == runtime.GOOS {
		java += `.exe`
	}

	return &Java{Home: javaHome, executable: java}
}
