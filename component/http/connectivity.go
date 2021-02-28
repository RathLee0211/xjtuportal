package http

import (
	"auto-portal-auth/component/basic"
	"auto-portal-auth/component/utils"
	"context"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type DnsHelper struct {
	loggerHelper *basic.LoggerHelper
	DnsSettings  *basic.ProgramDnsSettings
}

func InitDnsHelper(
	configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
) (*DnsHelper, error) {

	if configHelper == nil {
		err := errors.New("ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("loggerHelper is invalid")
		return nil, err
	}

	dnsHelper := &DnsHelper{
		loggerHelper: loggerHelper,
		DnsSettings:  &configHelper.ProgramSettings.ProgramDnsSettings,
	}

	return dnsHelper, nil

}

func (dnsHelper *DnsHelper) LookupCheck(domain string) (err error) {

	timeout := time.Duration(dnsHelper.DnsSettings.Connect.Timeout) * time.Second

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel() // important to avoid a resource leak

	dnsHelper.loggerHelper.AddLog(basic.DEBUG,
		fmt.Sprintf("http/connectivity: Lookup host for domain [%s]", domain))

	var r net.Resolver
	ips, err := r.LookupHost(ctx, domain)
	if err != nil {
		return
	}

	if len(ips) > 0 {
		dnsHelper.loggerHelper.AddLog(basic.DEBUG,
			fmt.Sprintf("http/connectivity: DNS lookup result\n%s", strings.Join(ips, "\n")))
	} else {
		dnsHelper.loggerHelper.AddLog(basic.WARNING, "http/connectivity: No result for lookup")
	}

	return

}

func (dnsHelper *DnsHelper) DnsCheck(domain string, server string) (err error) {
	query := new(dns.Msg)
	query.Id = dns.Id()
	query.RecursionDesired = true
	query.Question = make([]dns.Question, 1)
	query.Question[0] = dns.Question{
		Name:   fmt.Sprintf("%s.", domain),
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	}
	client := new(dns.Client)
	client.Dialer = &net.Dialer{
		Timeout: time.Duration(dnsHelper.DnsSettings.Connect.Timeout) * time.Second,
	}
	dnsHelper.loggerHelper.AddLog(basic.DEBUG,
		fmt.Sprintf("http/connectivity: Send type A query to DNS server %s", server))
	regex := regexp.MustCompile(`:\d{1,5}$`)
	if !regex.MatchString(server) {
		server = fmt.Sprintf("%s:53", server)
	}
	in, _, err := client.Exchange(query, server)

	if err != nil { // Query with error
		return err
	}

	if in.Rcode != dns.RcodeSuccess {
		err = errors.New(fmt.Sprintf("http/connectivity: Response from %s with error %d", server, in.Rcode))
		return err
	}

	if len(in.Answer) == 0 {
		err = errors.New("empty answer")
		return err
	}

	results := make([]string, 0, len(in.Answer))
	for _, answer := range in.Answer {
		if result, ok := answer.(*dns.A); ok {
			results = append(results, result.String())
		}
	}
	if len(results) == 0 {
		err = errors.New(fmt.Sprintf("http/connectivity: cannot get any DNS A record from %s", server))
		return err
	}

	dnsHelper.loggerHelper.AddLog(basic.DEBUG,
		fmt.Sprintf("http/connectivity: Query result from Server [%s]:\n%s", server, strings.Join(results, "\n")))
	return nil

}

type ConnectivityChecker struct {
	loggerHelper         *basic.LoggerHelper
	requestHelper        *RequestHelper
	dnsHelper            *DnsHelper
	connectivitySettings *basic.ProgramConnectivitySettings
}

func InitConnectivityChecker(
	configHelper *basic.ConfigHelper,
	loggerHelper *basic.LoggerHelper,
	requestHelper *RequestHelper,
	dnsHelper *DnsHelper,
) (*ConnectivityChecker, error) {

	if configHelper == nil {
		err := errors.New("ConfigHelper is invalid")
		return nil, err
	}

	if loggerHelper == nil {
		err := errors.New("LoggerHelper is invalid")
		return nil, err
	}

	if requestHelper == nil {
		err := errors.New("RequestHelper is invalid")
		return nil, err
	}

	if dnsHelper == nil {
		err := errors.New("DnsHelper is invalid")
		return nil, err
	}

	connectivityChecker := &ConnectivityChecker{
		loggerHelper:         loggerHelper,
		requestHelper:        requestHelper,
		dnsHelper:            dnsHelper,
		connectivitySettings: &configHelper.ProgramSettings.ProgramConnectivitySettings,
	}
	return connectivityChecker, nil

}

func (connectivityChecker *ConnectivityChecker) httpCheck(url string) (statusCode int, err error) {
	_, _, statusCode, err = connectivityChecker.requestHelper.SendRequest(
		url,
		"GET",
		nil,
		nil,
		make([]*http.Cookie, 0, 0),
	)
	if statusCode >= 300 {
		err = errors.New(fmt.Sprintf("response return error code [%d]", statusCode))
		return
	}
	return
}

func (connectivityChecker *ConnectivityChecker) IntranetHttpCheck() (statusCode int, err error) {
	statusCode, err = connectivityChecker.httpCheck(connectivityChecker.connectivitySettings.Http.Intranet)
	return
}

func (connectivityChecker *ConnectivityChecker) InternetHttpCheck() (statusCode int, err error) {
	statusCode, err = connectivityChecker.httpCheck(connectivityChecker.connectivitySettings.Http.Internet)
	return
}

func (connectivityChecker *ConnectivityChecker) DnsGroupCheck(
	serverGroup []string,
	domainGroup []string,
) (
	available []string,
	unavailable []string,
) {

	available = make([]string, len(serverGroup), len(serverGroup))
	unavailable = make([]string, len(serverGroup), len(serverGroup))
	wg := &sync.WaitGroup{}

	for index, server := range serverGroup {
		wg.Add(1)
		go func(threadId int, server string) {
			isAvailable := true
			for _, domain := range domainGroup {
				err := connectivityChecker.dnsHelper.DnsCheck(domain, server)
				if err != nil {
					connectivityChecker.loggerHelper.AddLog(basic.DEBUG, fmt.Sprintf("%v", err))
					isAvailable = false
					break
				}
				time.Sleep(time.Duration(connectivityChecker.dnsHelper.DnsSettings.Connect.Timeout) * time.Second)
			}
			if isAvailable {
				available[threadId] = server
				unavailable[threadId] = ""
			} else {
				available[threadId] = ""
				unavailable[threadId] = server
			}
			wg.Done()
		}(index, server)

	}
	wg.Wait()
	return utils.DeleteEmptyString(available), utils.DeleteEmptyString(unavailable)
}

func (connectivityChecker *ConnectivityChecker) IntranetDnsCheck() (available []string, unavailable []string) {
	domainList := []string{
		connectivityChecker.connectivitySettings.Dns.Domain.Intranet,
		connectivityChecker.connectivitySettings.Dns.Domain.Internet,
	}
	return connectivityChecker.DnsGroupCheck(connectivityChecker.connectivitySettings.Dns.Server.Intranet, domainList)
}

func (connectivityChecker *ConnectivityChecker) InternetDnsCheck() (available []string, unavailable []string) {
	domainList := []string{
		connectivityChecker.connectivitySettings.Dns.Domain.Intranet,
		connectivityChecker.connectivitySettings.Dns.Domain.Internet,
	}
	return connectivityChecker.DnsGroupCheck(connectivityChecker.connectivitySettings.Dns.Server.Internet, domainList)
}

func (connectivityChecker *ConnectivityChecker) SystemResolveCheck() (statusCode int, err error) {
	domainGroup := []string{
		connectivityChecker.connectivitySettings.Dns.Domain.Intranet,
		connectivityChecker.connectivitySettings.Dns.Domain.Internet,
	}
	for _, domain := range domainGroup {
		err = connectivityChecker.dnsHelper.LookupCheck(domain)
		if err != nil {
			return -1, err
		}
	}
	return 200, nil
}
