package test

import (
	"fmt"
	"strings"
	"testing"
	"xjtuportal/component/basic"
	"xjtuportal/component/device"
)

var (
	ifHelper = device.InterfaceHelper{}
)

func TestMacStandardize(t *testing.T) {

	// Test 0: Mac standardize
	macStandard := "00:00:5e:00:53:0a"
	macTest0 := []string{
		"00-00-5e-00-53-0a",
		"00-00-5E-00-53-0A",
		"00:00:5e:00:53:0a",
		"00:00:5E:00:53:0A",
		"00:00:5e:00:53:0A",
	}

	for _, mac := range macTest0 {
		if standard, err := device.MacStandardize(mac); err != nil || standard != macStandard {
			t.Error("Cannot standardize MAC address " + mac)
		}
	}

	// Test 1: error handling
	macTest1 := []string{
		"",
		",*^2c",
		"1234",
		"00:00:5e:00:53:0aAC",
		"hijk",
		"中文测试",
	}

	for _, mac := range macTest1 {
		if _, err := device.MacStandardize(mac); err == nil {
			t.Error("Cannot handle error MAC address " + mac)
		}
	}

}

func TestGetLocalInterfaceMac(t *testing.T) {

	ifList, macList, ipList, err := device.GetLocalInterfaceInfo()
	if err != nil {
		t.Error("Cannot get interface(s)")
	}
	for _, i := range ifList {
		fmt.Println(i)
	}
	fmt.Println(fmt.Sprintf(
		"MAC address(es) of local interface(s):\n%s",
		strings.Join(macList, ",\n"),
	))
	fmt.Println(fmt.Sprintf(
		"IP address(es) of local interface(s):\n%s",
		strings.Join(ipList, ",\n"),
	))
}

func TestFindLogoutMac(t *testing.T) {

	configHelper, loggerHelper, err := readConfig()
	if err != nil {
		basic.LoggerTemp.AddLog(basic.FATAL, fmt.Sprintf("%v", err))
		t.Error("Initialization ConfigHelper & LoggerHelper failed")
		return
	}
	interfaceHelper, err := device.InitInterfaceHelper(configHelper, loggerHelper)
	if err != nil {
		basic.LoggerTemp.AddLog(basic.ERROR, fmt.Sprintf("%v", err))
		t.Error("Initialization InterfaceHelper failed")
		return
	}
	// Test 0: Find an unknown MAC address at bottom
	sessionMacList0 := []string{
		"11:22:33:44:55:66",
		"aa:bb:cc:dd:ee:ff",
		"00:00:5e:00:53:01",
	}
	logoutMac := "00:00:5e:00:53:01"
	if interfaceHelper.FindLogoutMac(sessionMacList0) != logoutMac {
		t.Error("Error when finding unknown MAC address")
	}

	// Test 1: Find an unknown MAC address at top
	sessionMacList1 := []string{
		"11:22:33:44:55:66",
		"00:00:5e:00:53:01",
		"aa:bb:cc:dd:ee:ff",
	}
	logoutMac = "00:00:5e:00:53:01"
	if interfaceHelper.FindLogoutMac(sessionMacList1) != logoutMac {
		t.Error("Error when finding unknown MAC address")
	}

	// Test 2: Find a known MAC address in config-defined MAC list
	sessionMacList2 := []string{
		"11:22:33:44:55:66",
		"aa:bb:cc:dd:ee:ff",
	}
	logoutMac = "11:22:33:44:55:66"
	if interfaceHelper.FindLogoutMac(sessionMacList2) != logoutMac {
		t.Error("Error when finding a MAC address in known MAC list")
	}

	// Test 3: Find a known MAC address in macListHelper list
	sessionMacList3 := []string{
		"00:50:56:c0:00:08",
	}
	logoutMac = "00:50:56:c0:00:08"
	if interfaceHelper.FindLogoutMac(sessionMacList3) != logoutMac {
		t.Error("Error when finding a MAC address in known MAC list")
	}

	// Test 4: Empty session list
	sessionMacList4 := make([]string, 0, 1)
	logoutMac = ""
	if interfaceHelper.FindLogoutMac(sessionMacList4) != logoutMac {
		t.Error("Error handling empty session list")
	}

	// Test 6: Session list with duplicated element
	sessionMacList5 := []string{
		"00:50:56:c0:00:08",
		"00:50:56:c0:00:08",
		"aa:bb:cc:dd:ee:ff",
	}
	logoutMac = "aa:bb:cc:dd:ee:ff"
	if interfaceHelper.FindLogoutMac(sessionMacList5) != logoutMac {
		t.Error("Error handling session list with duplicated element")
	}

}
