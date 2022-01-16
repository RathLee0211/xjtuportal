package http

import (
	"fmt"
	"testing"
	"xjtuportal/component/basic"
)

var (
	conf = map[string][]string{
		"1080:1083":   {"shadowsocks", "shadowsocksR", "privoxy"},
		"2800:2803":   {"Netch"},
		"7890:7893":   {"clash", "Clash for Windows"},
		"8080:8088":   {},
		"10808:10810": {"v2rayN"},
	}
)

func TestProxyCheck(t *testing.T) {
	conf = map[string][]string{
		"1080:1083":   {"shadowsocks", "shadowsocksR", "privoxy"},
		"2800:2803":   {"Netch"},
		"7890:7893":   {"clash", "Clash for Windows"},
		"8080:8088":   {},
		"10808:10810": {"v2rayN"},
	}
	p := InitProxyHelper(basic.LoggerTemp, conf)
	fmt.Println(p.ProxyCheck())
	//fmt.Println(UrlConnCheck(testUrl, "http://127.0.0.1:7890", 1*time.Second))
}
