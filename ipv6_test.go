// go test -v -run="TestIPv6"

package cidrman

import (
	"net"
	"testing"
)

func TestIPv6(t *testing.T) {
	type TestCase struct {
		Input     string
		Netmask   string
		Broadcast string
		Network   string
		Error     bool
	}

	testCases := []TestCase{
		{
			Input:     "",
			Netmask:   "",
			Broadcast: "",
			Network:   "",
			Error:     true,
		},
		{
			Input:     "fe80::dead:beef/64",
			Netmask:   "ffff:ffff:ffff:ffff::",
			Broadcast: "fe80::ffff:ffff:ffff:ffff",
			Network:   "fe80::",
			Error:     false,
		},
	}

	for _, testCase := range testCases {
		ip, net, err := net.ParseCIDR(testCase.Input)
		if err != nil {
			if !testCase.Error {
				t.Errorf("net.ParseCIDR(%#v) failed: %s", testCase.Input, err.Error())
			}
			continue
		}

		prefix, _ := net.Mask.Size()

		netmask := uint128ToIPV6(netmask6(uint(prefix))).String()
		if netmask != testCase.Netmask {
			t.Errorf("Netmask expected: %#v, got: %#v", testCase.Netmask, netmask)
		}

		addr := ipv6ToUInt128(ip.To16())

		broadcast := uint128ToIPV6(broadcast6(addr, uint(prefix))).String()
		if broadcast != testCase.Broadcast {
			t.Errorf("Broadcast expected: %#v, got: %#v", testCase.Broadcast, broadcast)
		}

		network := uint128ToIPV6(network6(addr, uint(prefix))).String()
		if network != testCase.Network {
			t.Errorf("Network expected: %#v, got: %#v", testCase.Network, network)
		}
	}
}
