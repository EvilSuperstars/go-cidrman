// Inspired by the Python netaddr cidr_merge function
// https://netaddr.readthedocs.io/en/latest/api.html#netaddr.cidr_merge.

package cidrmerge

import (
	"fmt"
	"math/big"
	"net"
	"sort"
)

type ipNets []*net.IPNet

func (nets ipNets) toCIDRs() []string {
	var cidrs []string
	for _, net := range nets {
		cidrs = append(cidrs, net.String())
	}

	return cidrs
}

// MergeIPNets accepts a list of IP networks and merges them into the smallest possible list of IPNets.
// It merges adjacent subnets where possible, those contained within others and removes any duplicates.
func MergeIPNets(networks []*net.IPNet) ([]*net.IPNet, error) {
	if networks == nil {
		return nil, nil
	}
	if len(networks) == 0 {
		return make([]*net.IPNet, 0), nil
	}

	var blocks cidrBlocks
	for _, network := range networks {
		blocks = append(blocks, newBlock(network))
	}

	sort.Sort(blocks)

	// Coalesce overlapping blocks.
	for i := len(blocks) - 1; i > 0; i-- {
		if blocks[i].isIPv4 == blocks[i-1].isIPv4 {
			if blocks[i].first.Cmp(blocks[i-1].last) < 0 {
				blocks[i-1].last = blocks[i].last
				if blocks[i].first.Cmp(blocks[i-1].first) < 0 {
					blocks[i-1].first = blocks[i].first
				}
				blocks[i-1].network = nil
				blocks[i] = nil
			}
		}
	}

	var merged []*net.IPNet
	for _, block := range blocks {
		if block == nil {
			continue
		}

		// If this block wasn't coalesced just used the passed in IP network.
		if block.network != nil {
			merged = append(merged, block.network)
		}

		merged = append(merged, ipRangeToNets(block.first, block.last)...)
	}

	return merged, nil
}

// MergeCIDRs accepts a list of CIDR blocks and merges them into the smallest possible list of CIDRs.
func MergeCIDRs(cidrs []string) ([]string, error) {
	if cidrs == nil {
		return nil, nil
	}
	if len(cidrs) == 0 {
		return make([]string, 0), nil
	}

	var networks []*net.IPNet
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	mergedNets, err := MergeIPNets(networks)
	if err != nil {
		return nil, err
	}

	return ipNets(mergedNets).toCIDRs(), nil
}

type cidrBlock struct {
	isIPv4  bool
	first   *big.Int
	last    *big.Int
	network *net.IPNet
}

type cidrBlocks []*cidrBlock

func newBlock(network *net.IPNet) *cidrBlock {
	var block cidrBlock

	block.network = network
	if len(network.IP) == net.IPv4len {
		block.isIPv4 = true
	}

	first, last := addressRange(network)
	block.first, _ = ipToInt(first)
	block.last, _ = ipToInt(last)

	return &block
}

func (c cidrBlocks) Len() int {
	return len(c)
}

func (c cidrBlocks) Less(i, j int) bool {
	lhs := c[i]
	rhs := c[j]

	// IPv4 before IPv6.
	if lhs.isIPv4 && !rhs.isIPv4 {
		return true
	}
	if rhs.isIPv4 && !lhs.isIPv4 {
		return false
	}

	// Then by last IP in the range.
	cmp := lhs.last.Cmp(rhs.last)
	switch cmp {
	case -1:
		return true
	case 1:
		return false
	}

	// Then by first IP in the range
	cmp = lhs.first.Cmp(rhs.first)
	switch cmp {
	case -1:
		return true
	case 1:
		return false
	}

	return false
}

func (c cidrBlocks) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func ipRangeToNets(first, last *big.Int) []*net.IPNet {
	return nil
}

// Lifted from github.com/apparentlymart/go-cidr/cidr/cidr.go.
// AddressRange returns the first and last addresses in the given CIDR range.
func addressRange(network *net.IPNet) (net.IP, net.IP) {
	// the first IP is easy
	firstIP := network.IP

	// the last IP is the network address OR NOT the mask address
	prefixLen, bits := network.Mask.Size()
	if prefixLen == bits {
		// Easy!
		// But make sure that our two slices are distinct, since they
		// would be in all other cases.
		lastIP := make([]byte, len(firstIP))
		copy(lastIP, firstIP)
		return firstIP, lastIP
	}

	firstIPInt, bits := ipToInt(firstIP)
	hostLen := uint(bits) - uint(prefixLen)
	lastIPInt := big.NewInt(1)
	lastIPInt.Lsh(lastIPInt, hostLen)
	lastIPInt.Sub(lastIPInt, big.NewInt(1))
	lastIPInt.Or(lastIPInt, firstIPInt)

	return firstIP, intToIP(lastIPInt, bits)
}

// Lifted from github.com/apparentlymart/go-cidr/cidr/wrangling.go.
func ipToInt(ip net.IP) (*big.Int, int) {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	if len(ip) == net.IPv4len {
		return val, 32
	} else if len(ip) == net.IPv6len {
		return val, 128
	} else {
		panic(fmt.Errorf("Unsupported address length %d", len(ip)))
	}
}

func intToIP(ipInt *big.Int, bits int) net.IP {
	ipBytes := ipInt.Bytes()
	ret := make([]byte, bits/8)
	// Pack our IP bytes into the end of the return array,
	// since big.Int.Bytes() removes front zero padding.
	for i := 1; i <= len(ipBytes); i++ {
		ret[len(ret)-i] = ipBytes[len(ipBytes)-i]
	}
	return net.IP(ret)
}
