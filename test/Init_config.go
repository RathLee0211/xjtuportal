package test

import (
	"auto-portal-auth/component/base"
)

func readConfig() (*base.ConfigHelper, *base.LoggerHelper, error) {

	configHelper, err := base.InitConfigHelper(
		"../config/user-settings.yaml",
		"../config/program-settings.yaml",
	)
	if err != nil {
		return nil, nil, err
	}

	loggerHelper, err := base.InitLoggerHelper(configHelper)
	if err != nil {
		return nil, nil, err
	}

	//fmt.Println(configHelper.UserSettings)
	//fmt.Println(configHelper.ProgramSettings)

	return configHelper, loggerHelper, nil
}
