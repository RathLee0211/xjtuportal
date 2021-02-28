package test

import (
	"auto-portal-auth/component/app"
	"auto-portal-auth/component/basic"
	"auto-portal-auth/component/http"
	"fmt"
	"testing"
)

func TestConnectivity(t *testing.T) {

	configHelper, loggerHelper, err := readConfig()
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		t.Error("Initialization ConfigHelper & LoggerHelper failed")
		return
	}

	requestHelper, err := http.InitRequestHelper(configHelper, loggerHelper)
	if err != nil {
		t.Error("Initialization RequestHelper failed")
		return
	}

	dnsHelper, err := http.InitDnsHelper(configHelper, loggerHelper)
	if err != nil {
		t.Error("Initialization DNSHelper failed")
		return
	}

	connectivityChecker, err := http.InitConnectivityChecker(configHelper, loggerHelper, requestHelper, dnsHelper)
	if err != nil {
		t.Error("Initialization connectivityChecker failed")
		return
	}

	diagnosisHelper, err := app.InitDiagnosisHelper(configHelper, loggerHelper, connectivityChecker)
	if err != nil {
		t.Error("Initialization DiagnosisShellHelper failed")
		return
	}

	diagnosisHelper.DoDiagnosis()

}
