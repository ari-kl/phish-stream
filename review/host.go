package review

import (
	"context"
	"errors"
	"net"

	"github.com/ammario/ipisp/v2"
)

func LookupISP(domain string) (error, string, string) {
	ip, err := net.LookupIP(domain)

	if err != nil {
		return err, "", ""
	}

	if len(ip) == 0 {
		return errors.New("No IP addresses found for domain"), "", ""
	}

	resp, err := ipisp.LookupIP(context.Background(), ip[0])

	if err != nil {
		return err, "", ""
	}

	return nil, resp.ISPName, resp.Country
}
