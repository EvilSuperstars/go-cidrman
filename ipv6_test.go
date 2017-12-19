// go test -v -run="TestIPv6"

package cidrman

import (
	"net"
	"testing"
)

func TestIPv6(t *testing.T) {
	type TestCase struct {
		Input     string
		Hostmask  string
		Netmask   string
		Broadcast string
		Network   string
		Error     bool
	}

	testCases := []TestCase{
		{
			Input:     "",
			Hostmask:  "",
			Netmask:   "",
			Broadcast: "",
			Network:   "",
			Error:     true,
		},
		{
			Input:     "::/0",
			Hostmask:  "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			Netmask:   "::",
			Broadcast: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			Network:   "::",
			Error:     false,
		},
		{
			Input:     "fe80::dead:beef/64",
			Hostmask:  "::ffff:ffff:ffff:ffff",
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

		hostmask := uint128ToIPV6(hostmask6(uint(prefix))).String()
		if hostmask != testCase.Hostmask {
			t.Errorf("Hostmask(%#v) expected: %#v, got: %#v", testCase.Input, testCase.Hostmask, hostmask)
		}

		netmask := uint128ToIPV6(netmask6(uint(prefix))).String()
		if netmask != testCase.Netmask {
			t.Errorf("Netmask(%#v) expected: %#v, got: %#v", testCase.Input, testCase.Netmask, netmask)
		}

		addr := ipv6ToUInt128(ip.To16())

		broadcast := uint128ToIPV6(broadcast6(addr, uint(prefix))).String()
		if broadcast != testCase.Broadcast {
			t.Errorf("Broadcast(%#v) expected: %#v, got: %#v", testCase.Input, testCase.Broadcast, broadcast)
		}

		network := uint128ToIPV6(network6(addr, uint(prefix))).String()
		if network != testCase.Network {
			t.Errorf("Network(%#v) expected: %#v, got: %#v", testCase.Input, testCase.Network, network)
		}
	}
}
