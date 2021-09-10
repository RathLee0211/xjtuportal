package device

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"xjtuportal/component/basic"
	"xjtuportal/component/utils"
)

func MacStandardize(mac string) (string, error) {
	if macAddr, err := net.ParseMAC(mac); err != nil {
		return "", err
	} else {
		return net.HardwareAddr.String(macAddr), err
	}
}

type InterfaceHelper struct {
	loggerHelper       *basic.LoggerHelper
	userDeviceSettings *basic.UserDeviceSettings
	KnownMacMap        map[string]struct{}
	KnownMacList       []string
	LocalMacList       []string
	LocalIpList        []string
}

func (ifHelper *InterfaceHelper) MacListStandardize(macList []string) (standardList []string, errorList []string) {

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

func (ifHelper *InterfaceHelper) GetLocalInterfaceInfo() (macList []string, ipList []string, err error) {
	interfaces, err := net.Interfaces()
	macList = make([]string, 0, len(interfaces))
	ipList = make([]string, 0, len(interfaces))
	if err != nil {
		return macList, ipList, err
	}
	for _, i := range interfaces {
		mac := i.HardwareAddr
		if mac != nil {
			macList = append(macList, net.HardwareAddr.String(mac))
		}
		addrList, err := i.Addrs()
		if err != nil {
			ifHelper.loggerHelper.AddLog(basic.WARNING,
				fmt.Sprintf(
					"device/interface: Cannot get IP address from interface [%s]",
					i.Name,
				))
		} else {
			fmt.Println(addrList)
		}
	}

	return macList, ipList, nil
}

func InitInterfaceHelper(configHelper *basic.ConfigHelper, loggerHelper *basic.LoggerHelper) (*InterfaceHelper, error) {

	if loggerHelper == nil {
		err := errors.New("device/interface: logger is invalid")
		return nil, err
	}

	if configHelper == nil {
		err := errors.New("device/interface: ConfigHelper is invalid")
		return nil, err
	}

	interfaceHelper := &InterfaceHelper{
		loggerHelper:       loggerHelper,
		userDeviceSettings: &configHelper.UserSettings.UserDeviceSettings,
		KnownMacList:       configHelper.UserSettings.UserDeviceSettings.KnownMacList,
	}

	// Standardize the MAC addresses from config
	knownMacList, errorMacList := interfaceHelper.MacListStandardize(interfaceHelper.KnownMacList)
	if len(errorMacList) > 0 {
		interfaceHelper.loggerHelper.AddLog(basic.WARNING,
			fmt.Sprintf(
				"device/interface: MAC address(es) with invalid format:\n%s",
				strings.Join(errorMacList, ",\n"),
			))
	}

	// Get MAC addresses & IP addresses from local interfaces
	localMacList, localIpList, err := interfaceHelper.GetLocalInterfaceInfo()
	interfaceHelper.LocalMacList = localMacList
	interfaceHelper.LocalIpList = localIpList
	if err != nil {
		interfaceHelper.loggerHelper.AddLog(basic.WARNING,
			fmt.Sprintf("device/interface: Error when getting interfaces [%v]", err))
	}
	if len(localMacList) > 0 {
		interfaceHelper.loggerHelper.AddLog(basic.DEBUG,
			fmt.Sprintf(
				"device/interface: MAC address(es) of local interface(s):\n%s",
				strings.Join(localMacList, ", "),
			))
	}
	if len(localIpList) > 0 {
		interfaceHelper.loggerHelper.AddLog(basic.DEBUG,
			fmt.Sprintf(
				"device/interface: IP address(es) of local interface(s):\n%s",
				strings.Join(localIpList, ", "),
			))
	}

	if interfaceHelper.userDeviceSettings.UseInterface {
		knownMacList = append(knownMacList, localMacList...)
	}

	// Merge and remove duplicate
	interfaceHelper.KnownMacList, interfaceHelper.KnownMacMap = utils.RemoveDuplicateStrings(knownMacList)

	return interfaceHelper, nil
}

func (ifHelper *InterfaceHelper) FindLogoutMac(sessionMacList []string) (mac string) {

	if len(sessionMacList) == 0 {
		return ""
	}

	sessionMacList, sessionMacMap := utils.RemoveDuplicateStrings(sessionMacList)

	// Find a MAC address that exist in session list but not in known MAC list
	for _, mac := range sessionMacList {
		if _, ok := ifHelper.KnownMacMap[mac]; !ok {
			return mac
		}
	}

	// Find a MAC address that has a smaller index in known MAC list and exist in session list
	for _, mac := range ifHelper.KnownMacList {
		if _, ok := sessionMacMap[mac]; ok {
			return mac
		}
	}

	// Finally, choose the first mac in session list
	return sessionMacList[0]

}
