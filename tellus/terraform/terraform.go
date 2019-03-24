// Package terraform runs terraform commands.
// It requires that terraform in present on the PATH.
package terraform

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// Plan runs terraform plan in the given directory and returns the resulting output as a string.
// If the run wasn't successful the second return argument will be false.
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

// Apply runs terraform apply in the given directory and returns the resulting output as a string.
// If the run wasn't successful the second return argument will be false.
func Apply(directory string) (string, bool) {
	if err := os.Chdir(string(directory)); err != nil {
		return err.Error(), false
	}
	if err := exec.Command("terraform", "init").Run(); err != nil {
		return "could not initialize terraform", false
	}
	planCmd := exec.Command("terraform", "apply", "-no-color")
	output := &strings.Builder{}
	planCmd.Stdout = output
	planCmd.Stderr = output
	err := planCmd.Run()
	if err != nil {
		log.Print(err.Error())
	}
	return output.String(), err == nil
}
