package device

import (
	"auto-portal-auth/component/base"
	"auto-portal-auth/component/utils"
	"errors"
	"fmt"
	"net"
	"strings"
)

func MacStandardize(mac string) (string, error) {
	if macAddr, err := net.ParseMAC(mac); err != nil {
		return "", err
	} else {
		return net.HardwareAddr.String(macAddr), err
	}
}

func MacListStandardize(macList []string) (standardList []string, errorList []string) {

	standardMacList := make([]string, 0, len(macList))
	errorMacList := make([]string, 0, len(macList))

	for _, mac := range macList {
		standardMac, err := MacStandardize(mac)
		if err != nil {
			errorMacList = append(errorMacList, mac)
		} else {
			standardMacList = append(standardMacList, standardMac)
		}

	}
	return standardMacList, errorMacList
}

func GetLocalInterfaceMac() ([]string, error) {
	interfaces, err := net.Interfaces()
	macs := make([]string, 0, len(interfaces))
	if err != nil {
		return macs, err
	}
	for _, i := range interfaces {
		mac := i.HardwareAddr
		if mac != nil {
			macs = append(macs, net.HardwareAddr.String(mac))
		}
	}

	return macs, nil
}

type MacListHelper struct {
	loggerHelper *base.LoggerHelper
	KnownMacMap  map[string]struct{}
	KnownMacList []string
}

func InitMacListHelper(configHelper *base.ConfigHelper, loggerHelper *base.LoggerHelper) (*MacListHelper, error) {

	if loggerHelper == nil {
		err := errors.New("logger is invalid")
		return nil, err
	}

	if configHelper == nil {
		err := errors.New("ConfigHelper is invalid")
		return nil, err
	}

	macListHelper := &MacListHelper{}
	macListHelper.loggerHelper = loggerHelper

	// Standardize the MAC addresses from config
	knownMacList, errorMacList := MacListStandardize(configHelper.UserSettings.SessionManage.DeviceList)
	if len(errorMacList) > 0 {
		macListHelper.loggerHelper.AddLog(base.WARNING,
			fmt.Sprintf(
				"MAC address(es) with invalid format:\n%s",
				strings.Join(errorMacList, ",\n"),
			))
	}

	// Get MAC addresses from local interfaces
	localMacList, err := GetLocalInterfaceMac()
	if err != nil {
		macListHelper.loggerHelper.AddLog(base.WARNING,
			fmt.Sprintf("Cannot get network interface(s) [%v]", err))
	} else {
		macListHelper.loggerHelper.AddLog(base.DEBUG,
			fmt.Sprintf(
				"MAC address(es) of local interface(s):\n%s",
				strings.Join(localMacList, ",\n"),
			))
	}

	// Merge and remove duplicate
	knownMacList = append(knownMacList, localMacList...)
	macListHelper.KnownMacList, macListHelper.KnownMacMap = utils.RemoveDuplicateStrings(knownMacList)

	return macListHelper, nil
}

func (device *MacListHelper) FindLogoutMac(sessionMacList []string) (mac string) {

	if len(sessionMacList) == 0 {
		return ""
	}

	sessionMacList, sessionMacMap := utils.RemoveDuplicateStrings(sessionMacList)

	// Find a MAC address that exist in session list but not in known MAC list
	for _, mac := range sessionMacList {
		if _, ok := device.KnownMacMap[mac]; !ok {
			return mac
		}
	}

	// Find a MAC address that has a smaller index in known MAC list and exist in session list
	for _, mac := range device.KnownMacList {
		if _, ok := sessionMacMap[mac]; ok {
			return mac
		}
	}

	// Finally, choose the first mac in session list
	return sessionMacList[0]

}
