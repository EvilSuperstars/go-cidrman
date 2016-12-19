// Inspired by the Python netaddr iprange_to_cidrs function:
// https://netaddr.readthedocs.io/en/latest/api.html#netaddr.iprange_to_cidrs.

package cidrmerge

import (
	"errors"
	"fmt"
	"net"
)

// IPRangeToIPNets accepts an arbitrary start and end IP address and returns a list of
// CIDR subnets that fit exactly between the boundaries of the two with no overlap.
func IPRangeToIPNets(start, end net.IP) ([]*net.IPNet, error) {
	start4 := start.To4()
	end4 := end.To4()

	if ((start4 == nil) && (end4 != nil)) || ((start4 != nil) && (end4 == nil)) {
		return nil, errors.New("Mismatched IP address types")
	}

	var cidrs []*net.IPNet

	if start4 != nil {
		lo := ipv4ToUInt32(start4)
		hi := ipv4ToUInt32(end4)
		if hi < lo {
			return nil, errors.New("End < Start")
		}

		splitRange4(0, 0, lo, hi, &cidrs)
	} else {
		return nil, errors.New("Not implemented")
	}

	return cidrs, nil
}

// IPRangeToCIDRs accepts an arbitrary start and end IP address and returns a list of
// CIDR subnets that fit exactly between the boundaries of the two with no overlap.
func IPRangeToCIDRs(start, end string) ([]string, error) {
	ipStart := net.ParseIP(start)
	if ipStart == nil {
		return nil, fmt.Errorf("Invalid IP address: %s", start)
	}
	ipEnd := net.ParseIP(end)
	if ipEnd == nil {
		return nil, fmt.Errorf("Invalid IP address: %s", end)
	}

	nets, err := IPRangeToIPNets(ipStart, ipEnd)
	if err != nil {
		return nil, err
	}

	return ipNets(nets).toCIDRs(), nil
}
