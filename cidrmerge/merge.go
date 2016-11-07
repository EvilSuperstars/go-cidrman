// CIDR merging.
// Inspired by the Python netaddr cidr_merge function (https://netaddr.readthedocs.io/en/latest/tutorial_01.html).

package cidrmerge

import (
	"math/big"
	"net"
	"sort"

	"github.com/apparentlymart/go-cidr/cidr"
)

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

	var mergedCIDRs []string
	for _, network := range mergedNets {
		mergedCIDRs = append(mergedCIDRs, network.String())
	}

	return mergedCIDRs, nil
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

	first, last := cidr.AddressRange(network)
	block.first = ipToInt(first)
	block.last = ipToInt(last)

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

// Lifted from github.com/apparentlymart/go-cidr/cidr/wrangling.go.
func ipToInt(ip net.IP) *big.Int {
	val := &big.Int{}
	val.SetBytes([]byte(ip))
	return val
}

func ipRangeToNets(first, last *big.Int) []*net.IPNet {
	return nil
}
