package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func Exec(
	command string, arguments []string, environment []string,
) (int, *bytes.Buffer, *bytes.Buffer, error) {
	if arguments == nil {
		arguments = []string{}
	}

	if environment == nil {
		environment = []string{}
	}

	_environment := os.Environ()
	for _, e := range environment {
		_environment = append(_environment, e)
	}

	var exitStatus int
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(command, arguments...)
	cmd.Env = _environment
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				log.Errorf("ExitError while running '%s':\n%s", cmd, exitErr.Stderr)
				exitStatus = status.ExitStatus()
				log.Errorf("'%s' exited with '%d'", cmd, exitStatus)
			} else {
				exitStatus = -1
			}
		}

		exitStatus = -1
		log.Errorf("Error while running '%s': %s", cmd, err)
		log.Errorf(
			"exit status: %d\nstdout:\n%s\nstderr:\n%s",
			exitStatus, stdout.String(), stderr.String(),
		)
		return exitStatus, &stdout, &stderr, err
	}

	exitStatus = 0
	log.Infof(
		"exit status: %d\nstdout:\n%s\nstderr:\n%s",
		exitStatus, stdout.String(), stderr.String(),
	)
	return exitStatus, &stdout, &stderr, nil
}
