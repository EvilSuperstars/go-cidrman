// go test -v -run="TestIPv4"

package cidrman

import (
	"net"
	"testing"
)

func TestIPv4(t *testing.T) {
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
			Input:     "192.0.2.0/24",
			Netmask:   "255.255.255.0",
			Broadcast: "192.0.2.255",
			Network:   "192.0.2.0",
			Error:     true,
		},
		{
			Input:     "192.0.3.112/22",
			Netmask:   "255.255.252.0",
			Broadcast: "192.0.3.255",
			Network:   "192.0.0.0",
			Error:     true,
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

		netmask := uint32ToIPV4(netmask4(uint(prefix))).String()
		if netmask != testCase.Netmask {
			t.Errorf("Netmask expected: %#v, got: %#v", testCase.Netmask, netmask)
		}

		addr := ipv4ToUInt32(ip.To4())

		broadcast := uint32ToIPV4(broadcast4(addr, uint(prefix))).String()
		if broadcast != testCase.Broadcast {
			t.Errorf("Broadcast expected: %#v, got: %#v", testCase.Broadcast, broadcast)
		}

		network := uint32ToIPV4(network4(addr, uint(prefix))).String()
		if network != testCase.Network {
			t.Errorf("Network expected: %#v, got: %#v", testCase.Network, network)
		}
	}
}
