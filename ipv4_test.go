// go test -v -run="TestIPv4"

package cidrman

import (
	"net"
	"testing"
)

func TestIPv4(t *testing.T) {
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
			Input:     "0.0.0.0/0",
			Hostmask:  "255.255.255.255",
			Netmask:   "0.0.0.0",
			Broadcast: "255.255.255.255",
			Network:   "0.0.0.0",
			Error:     false,
		},
		{
			Input:     "192.0.2.0/24",
			Hostmask:  "0.0.0.255",
			Netmask:   "255.255.255.0",
			Broadcast: "192.0.2.255",
			Network:   "192.0.2.0",
			Error:     false,
		},
		{
			Input:     "192.0.3.112/22",
			Hostmask:  "0.0.3.255",
			Netmask:   "255.255.252.0",
			Broadcast: "192.0.3.255",
			Network:   "192.0.0.0",
			Error:     false,
		},
		{
			Input:     "192.168.1.151/32",
			Hostmask:  "0.0.0.0",
			Netmask:   "255.255.255.255",
			Broadcast: "192.168.1.151",
			Network:   "192.168.1.151",
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

		hostmask := uint32ToIPV4(hostmask4(uint(prefix))).String()
		if hostmask != testCase.Hostmask {
			t.Errorf("Hostmask expected: %#v, got: %#v", testCase.Hostmask, hostmask)
		}

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
