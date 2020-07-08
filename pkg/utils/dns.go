package utils

import (
	"fmt"
	"github.com/miekg/dns"
)

var publicDnsServers = []string{"8.8.8.8:53", "8.8.4.4:53"}

// ResolveName returns public IP address associated with the provided
// DNS record
func ResolveName(name string, version int) ([]string, error) {
	addrs := []string{}
	if version != 4 {
		return addrs, fmt.Errorf("only ip version 4 is supported")
	}

	for _, server := range publicDnsServers {
		req := new(dns.Msg)
		req.Id = dns.Id()
		req.RecursionDesired = true
		req.Question = make([]dns.Question, 1)
		req.Question[0] = dns.Question{dns.Fqdn(name), dns.TypeA, dns.ClassINET}
		resp, err := dns.Exchange(req, server)
		if err != nil {
			return addrs, err
		}

		if resp != nil && resp.Rcode != dns.RcodeSuccess {
			return addrs, fmt.Errorf("%s", dns.RcodeToString[resp.Rcode])
		}

		for _, record := range resp.Answer {
			if t, ok := record.(*dns.A); ok {
				addrs = append(addrs, t.A.String())
			}
		}
		if len(addrs) > 0 {
			return addrs, nil
		}
	}

	return addrs, nil
}
