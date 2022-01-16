package http

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"xjtuportal/component/basic"
	"xjtuportal/component/utils"
)

const (
	testUrl   = "http://www.gstatic.com/generate_204"
	localhost = "127.0.0.1"
)

var (
	prefixes = []string{
		"socks5://",
		"socks4://",
		"http://",
	}
)

type ProxyHelper struct {
	loggerHelper *basic.LoggerHelper
	proxyPorts   map[int][]string // [scheme]://[host]:[port]
	testUrl      string
	timeout      time.Duration
}

func InitProxyHelper(lh *basic.LoggerHelper, setting *basic.ConfigHelper) *ProxyHelper {
	proxyPorts := make(map[int][]string)
out:
	for k, v := range setting.ProgramSettings.ProgramConnectivitySettings.Proxy.Ports {
		for {

			if strings.IndexByte(k, ':') > -1 {
				portRange := strings.Split(k, ":")
				if len(portRange) != 2 {
					continue out
				}
				start, err := strconv.Atoi(portRange[0])
				if err != nil || start < 0 || start > 65535 {
					continue out
				}
				end, err := strconv.Atoi(portRange[1])
				if err != nil || start < 0 || start > 65535 || end < start {
					continue out
				}
				for i := start; i <= end; i++ {
					proxyPorts[i] = append(proxyPorts[i], v...)
				}
				continue out
			}

			if port, err := strconv.Atoi(k); err != nil && port >= 0 && port <= 65535 {
				proxyPorts[port] = append(proxyPorts[port], v...)
			}

			break
		}
	}
	for k := range proxyPorts {
		proxyPorts[k], _ = utils.RemoveDuplicateStrings(proxyPorts[k])
	}

	if _, err := url.ParseRequestURI(setting.ProgramSettings.ProgramConnectivitySettings.Proxy.TestUrl); err != nil {
		setting.ProgramSettings.ProgramConnectivitySettings.Proxy.TestUrl = testUrl
	}

	if setting.ProgramSettings.ProgramConnectivitySettings.Proxy.Timeout < 0 {
		setting.ProgramSettings.ProgramConnectivitySettings.Proxy.Timeout = 2
	}
	return &ProxyHelper{
		testUrl:      setting.ProgramSettings.ProgramConnectivitySettings.Proxy.TestUrl,
		timeout:      time.Duration(setting.ProgramSettings.ProgramConnectivitySettings.Proxy.Timeout) * time.Second,
		loggerHelper: lh,
		proxyPorts:   proxyPorts,
	}
}

func (ph *ProxyHelper) ProxyCheck() (proxies []string, programs []string, available []bool) {
	var wg sync.WaitGroup
	i := 0
	proxies = make([]string, len(ph.proxyPorts)*len(prefixes))
	ports := make([]int, len(ph.proxyPorts)*len(prefixes))
	available = make([]bool, len(ph.proxyPorts)*len(prefixes))
	for p := range ph.proxyPorts {
		port := p
		proxy := fmt.Sprint(localhost, ":", port)
		for j := 0; j < len(prefixes); j++ {
			checkIndex := i*len(prefixes) + j
			ports[checkIndex] = port
			wg.Add(1)
			go func(index int, proxyUrlStr string) {
				defer wg.Done()
				if !proxyExist(proxyUrlStr, ph.timeout) {
					return
				}
				proxies[index] = proxyUrlStr
				if UrlConnCheck(ph.testUrl, proxyUrlStr, ph.timeout) {
					available[index] = true
				}
			}(checkIndex, fmt.Sprint(prefixes[j], proxy))
		}
		i++
	}
	wg.Wait()
	c := 0
	for i = 0; i < len(proxies); i++ {
		if proxies[i] != "" {
			proxies[c] = proxies[i]
			programs = append(programs, strings.Join(ph.proxyPorts[ports[i]], ", "))
			available[c] = available[i]
			c++
		}
	}
	proxies = proxies[:c]
	available = available[:c]
	return

}

func proxyExist(proxyUrlStr string, timeout time.Duration) bool {
	proxyUrl, err := url.ParseRequestURI(proxyUrlStr)
	if err != nil {
		return false
	}
	proxyHost := proxyUrl.Host
	if _, _, err = net.SplitHostPort(proxyHost); err != nil {
		return false
	}

	switch strings.ToLower(proxyUrl.Scheme) {
	case "socks5", "socks":
		return socks5ProxyExist(proxyHost, timeout)
	case "socks4":
		return socks4ProxyExit(proxyHost, timeout)
	case "http":
		return httpProxyExist(proxyHost, timeout)
	default:
		return false
	}
}

func httpProxyExist(proxyHost string, timeout time.Duration) bool {
	hd, err := http.NewRequest(http.MethodConnect, fmt.Sprint("http://", proxyHost), nil)
	if err != nil {
		return false
	}
	client := &http.Client{
		Timeout: timeout,
	}
	defer client.CloseIdleConnections()
	resp, err := client.Do(hd)
	if err != nil || resp == nil {
		return false
	}
	return true
}

func socks5ProxyExist(proxyHost string, timeout time.Duration) bool {
	return socksProxyExist(5, proxyHost, timeout)
}

func socks4ProxyExit(proxyHost string, timeout time.Duration) bool {
	return socksProxyExist(4, proxyHost, timeout)
}

func socksProxyExist(version int, proxyHost string, timeout time.Duration) bool {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", proxyHost)
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	err = conn.SetReadDeadline(time.Now().Add(timeout * 2))
	if err != nil {
		return false
	}

	var data []byte
	var flag int
	var flagIndex int

	switch version {
	case 4:
		data = []byte{0x04, 0x01, 0x00, 0x50, 0x42, 0x66, 0x07, 0x63, 0x46, 0x72, 0x65, 0x64, 0x00}
		flag = 0x5a
		flagIndex = 1
	case 5:
		data = []byte{0x05, 0x01, 0x00}
		flag = 5
		flagIndex = 0
	default:
		return false
	}

	_, err = conn.Write(data)
	if err != nil {
		return false
	}
	recv := make([]byte, 2)
	_, err = conn.Read(recv)
	if err != nil {
		return false
	}
	if int(recv[flagIndex]) == flag {
		return true
	}
	return false

}

func UrlConnCheck(tUrl string, proxyUrl string, timeout time.Duration) bool {

	if _, err := url.ParseRequestURI(tUrl); err != nil {
		return false
	}
	hd, err := http.NewRequest(http.MethodHead, tUrl, nil)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{},
	}

	client.Transport.(*http.Transport).Proxy = func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}

	resp, err := client.Do(hd)
	if err != nil || resp == nil || resp.StatusCode >= http.StatusMultipleChoices {
		return false
	}
	return true

}
