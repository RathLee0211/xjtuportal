package main

import (
	"auto-portal-auth/exec"
)

func main() {
	shellRun := exec.InitShellUi()
	if shellRun != nil {
		shellRun.Exec()
	}
}
