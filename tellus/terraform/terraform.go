package terraform

import (
	"os"
	"os/exec"
	"strings"
)


func Plan(directory string) (string, bool) {
	if err := os.Chdir(string(directory)); err != nil {
		return "", false
	}
	if err := exec.Command("terraform", "init").Run(); err != nil {
		return "could not initialize terraform", false
	}
	planCmd := exec.Command("terraform", "plan", "-no-color")
	output := &strings.Builder{}
	planCmd.Stdout = output
	planCmd.Stderr = output
	err := planCmd.Run()
	return  output.String(), err == nil
}

func Apply(directory string) (string, bool) {
	if err := os.Chdir(string(directory)); err != nil {
		return err.Error(), false
	}
	if err := exec.Command("terraform", "init").Run(); err != nil {
		return "could not initialize terraform", false
	}
	planCmd := exec.Command("terraform", "apply", "-auto-approve")
	output := &strings.Builder{}
	planCmd.Stdout = output
	err := planCmd.Run()
	return output.String(), err == nil
}
