package netutil

import (
	"net"

	"github.com/askasoft/pango/str"
)

func ParseCIDRs(cidr string) (cidrs []*net.IPNet) {
	ss := str.Fields(cidr)
	for _, s := range ss {
		_, cidr, err := net.ParseCIDR(s)
		if err == nil {
			cidrs = append(cidrs, cidr)
		}
	}
	return
}
