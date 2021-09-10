package test

import (
	"xjtuportal/component/basic"
)

func readConfig() (*basic.ConfigHelper, *basic.LoggerHelper, error) {

	configHelper, err := basic.InitConfigHelper(
		"../config/user-settings.yaml",
		"../config/program-settings.yaml",
	)
	if err != nil {
		return nil, nil, err
	}

	loggerHelper, err := basic.InitLoggerHelper(configHelper)
	if err != nil {
		return nil, nil, err
	}

	//fmt.Println(configHelper.UserSettings)
	//fmt.Println(configHelper.ProgramSettings)

	return configHelper, loggerHelper, nil
}
