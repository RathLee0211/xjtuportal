package main

import (
	"flag"
	"fmt"
	"xjtuportal/component/utils"
	"xjtuportal/exec"
)

func main() {

	currentRunningDir := utils.GetCurrentRunningDir()

	versionFlag := flag.Bool("v", false, "Show current version")
	configFlag := flag.String("c", fmt.Sprintf("%s/%s", currentRunningDir, "config"), "The path of config folder")
	loginFlag := flag.Bool("i", false, "Login using auth data given in config file")
	logoutFlag := flag.Int("o", -1, "Logout with given index (shown by -s)")
	showSessionFlag := flag.Bool("s", false, "List current sessions")
	diagnosisFlag := flag.Bool("d", false, "Check http and DNS connectivity")
	adapterFlag := flag.Bool("a", false, "Check network adapter information")

	flag.Parse()

	for {
		shellRun := exec.InitShellUi(
			*versionFlag,
			*configFlag,
			*loginFlag,
			*logoutFlag,
			*showSessionFlag,
			*diagnosisFlag,
			*adapterFlag,
		)
		if shellRun != nil {
			exit := shellRun.Exec()
			if exit {
				break
			}
		} else {
			break
		}
	}

}
