package device

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"xjtuportal/component/basic"
	"xjtuportal/component/utils"
)

const (
	PrefixIpVersion = 0x19140000 + iota
	Ipv4
	Ipv6
)

var (
	NotValidIpChar = regexp.MustCompile(`[^a-fA-F0-9.:/]`)
)

func MacStandardize(mac string) (string, error) {
	if macAddr, err := net.ParseMAC(mac); err != nil {
		return "", err
	} else {
		return net.HardwareAddr.String(macAddr), err
	}
}

func IpVersionCheck(ip interface{}) (versionNum int) {
	versionNum = 0
	switch ipType := ip.(type) {
	case net.IP:
		if ipType.To4() != nil {
			versionNum = Ipv4
		} else {
			versionNum = Ipv6
		}
	case string:
		for i := 0; i < len(ipType); i++ {
			switch ipType[i] {
			case '.':
				versionNum = Ipv4
			case ':':
				versionNum = Ipv6
			}
		}
	}
	return versionNum
}

func ParseCidr(cidrStr string) (net.IP, *net.IPNet) {

	cidrStr = NotValidIpChar.ReplaceAllString(cidrStr, "")
	isMaskExist := false
	for i := 0; i < len(cidrStr); i++ {
		if cidrStr[i] == '/' {
			isMaskExist = true
			break
		}
	}

	versionNum := IpVersionCheck(cidrStr)
	if !isMaskExist {
		switch versionNum {
		case Ipv4:
			cidrStr = utils.Sprint(cidrStr, "/32")
		case Ipv6:
			cidrStr = utils.Sprint(cidrStr, "/128")
		default:
			return nil, nil
		}
	}

	ip, ipNet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, nil
	}

	return ip, ipNet
}

func GetLocalInterfaceInfo() (ifList []*InterfaceInfo, macList, ipList []string, err error) {
	interfaces, err := net.Interfaces()

	ifList = make([]*InterfaceInfo, 0, len(interfaces))
	macList = make([]string, 0, len(interfaces))
	ipList = make([]string, 0, len(interfaces))

	if err != nil {
		return nil, nil, nil, err
	}

	_, localCidr, _ := net.ParseCIDR("127.0.0.0/8")
	_, internalCidr, _ := net.ParseCIDR("169.254.0.0/16")

	for _, i := range interfaces {

		var macStr string
		if mac := i.HardwareAddr; mac != nil {
			macStr = net.HardwareAddr.String(mac)
		}

		curIp := make([]string, 0)
		filteredIp := make([]string, 0)

		if addrList, err := i.Addrs(); err == nil {
			for _, ipAddr := range addrList {
				ip, _ := ParseCidr(ipAddr.String())
				curIp = append(curIp, ip.String())
				if IpVersionCheck(ip) == Ipv6 || localCidr.Contains(ip) || internalCidr.Contains(ip) {
					continue
				}
				filteredIp = append(filteredIp, ip.String())
			}
		}

		ipList = append(ipList, filteredIp...)
		if len(filteredIp) > 0 {
			macList = append(macList, macStr)
		}

		ifList = append(ifList, &InterfaceInfo{
			name:   i.Name,
			mac:    macStr,
			ipList: curIp,
		})

	}

	return ifList, macList, ipList, nil
}

type InterfaceInfo struct {
	name   string
	mac    string
	ipList []string
}

func (i *InterfaceInfo) String() string {
	return utils.Sprint(i.name, "\n======================\n", "Mac: ", i.mac, "\n", "Address(es): ", i.ipList, "\n======================\n")
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
	_, localMacList, localIpList, err := GetLocalInterfaceInfo()
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
