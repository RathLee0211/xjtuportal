package main

import (
	"xjtuportal/exec"
)

func main() {
	shellRun := exec.InitShellUi()
	if shellRun != nil {
		shellRun.Exec()
	}
}
