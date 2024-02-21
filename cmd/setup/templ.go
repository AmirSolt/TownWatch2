package main

import (
	"fmt"
	"os/exec"
)

func templGenerate() {
	cmdName := "templ"
	args := []string{"generate"}

	cmd := exec.Command(cmdName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Output:\n%s\n", output)
}
