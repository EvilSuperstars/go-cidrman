package cidrman

import (
	"fmt"
	"math/big"
	"net"
)

const widthUInt128 = 128

// ipv6ToUInt128 converts an IPv6 address to an unsigned 128-bit integer.
func ipv6ToUInt128(ip net.IP) *big.Int {
	return big.NewInt(0).SetBytes(ip)
}

// uint128ToIPV6 converts an unsigned 128-bit integer to an IPv6 address.
func uint128ToIPV6(addr *big.Int) net.IP {
	return net.IP(addr.Bytes()).To16()
}

// copyUInt128 copies an unsigned 128-bit integer.
func copyUInt128(x *big.Int) *big.Int {
	return big.NewInt(0).Set(x)
}

// hostmask6 returns the hostmask for the specified prefix.
func hostmask6(prefix uint) *big.Int {
	z := big.NewInt(0)

	z.Lsh(big.NewInt(1), widthUInt128-prefix)
	z.Sub(z, big.NewInt(1))

	return z
}

// broadcast6 returns the broadcast address for the given address and prefix.
func broadcast6(addr *big.Int, prefix uint) *big.Int {
	z := big.NewInt(0)

	z.Or(addr, hostmask6(prefix))

	return z
}

// network6 returns the network address for the given address and prefix.
func network6(addr *big.Int, prefix uint) *big.Int {
	z := copyUInt128(addr)

	if prefix == 0 {
		return z
	}

	for i := int(prefix); i < 8*net.IPv6len; i++ {
		z = z.SetBit(z, i, 0)
	}
	return z
}

// splitRange6 recursively computes the CIDR blocks to cover the range lo to hi.
func splitRange6(addr *big.Int, prefix uint, lo, hi *big.Int, cidrs *[]*net.IPNet) error {
	if prefix > widthUInt128 {
		return fmt.Errorf("Invalid mask size: %d", prefix)
	}

	bc := broadcast6(addr, prefix)
	fmt.Printf("%v/%v, %v-%v, %v\n", uint128ToIPV6(addr), prefix, uint128ToIPV6(lo), uint128ToIPV6(hi), uint128ToIPV6(bc))
	if (lo.Cmp(addr) < 0) || (hi.Cmp(bc) > 0) {
		return fmt.Errorf("%v, %v out of range for network %v/%d, broadcast %v", uint128ToIPV6(lo), uint128ToIPV6(hi), uint128ToIPV6(addr), prefix, uint128ToIPV6(bc))
	}

	if (lo.Cmp(addr) == 0) && (hi.Cmp(bc) == 0) {
		cidr := net.IPNet{IP: uint128ToIPV6(addr), Mask: net.CIDRMask(int(prefix), 8*net.IPv6len)}
		*cidrs = append(*cidrs, &cidr)
		return nil
	}

	prefix++
	lowerHalf := copyUInt128(addr)
	upperHalf := copyUInt128(addr)
	upperHalf.SetBit(upperHalf, int(widthUInt128 - prefix), 1)
	if hi.Cmp(upperHalf) < 0 {
		return splitRange6(lowerHalf, prefix, lo, hi, cidrs)
	} else if lo.Cmp(upperHalf) >= 0 {
		return splitRange6(upperHalf, prefix, lo, hi, cidrs)
	} else {
		err := splitRange6(lowerHalf, prefix, lo, broadcast6(lowerHalf, prefix), cidrs)
		if err != nil {
			return err
		}
		return splitRange6(upperHalf, prefix, upperHalf, hi, cidrs)
	}
}

// IPv6 CIDR block.

type cidrBlock6 struct {
	first *big.Int
	last  *big.Int
}

type cidrBlock6s []*cidrBlock6

// newBlock6 returns a new IPv6 CIDR block.
func newBlock6(ip net.IP, mask net.IPMask) *cidrBlock6 {
	var block cidrBlock6

	block.first = ipv6ToUInt128(ip)
	prefix, _ := mask.Size()
	block.last = broadcast6(block.first, uint(prefix))

	return &block
}

// Sort interface.

func (c cidrBlock6s) Len() int {
	return len(c)
}

func (c cidrBlock6s) Less(i, j int) bool {
	lhs := c[i]
	rhs := c[j]

	// By last IP in the range.
	if lhs.last.Cmp(rhs.last) < 0 {
		return true
	} else if lhs.last.Cmp(rhs.last) > 0 {
		return false
	}

	// Then by first IP in the range.
	if lhs.first.Cmp(rhs.first) < 0 {
		return true
	} else if lhs.first.Cmp(rhs.first) > 0 {
		return false
	}

	return false
}

func (c cidrBlock6s) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// merge6 accepts a list of IPv6 networks and merges them into the smallest possible list of IPNets.
// It merges adjacent subnets where possible, those contained within others and removes any duplicates.
func merge6(blocks cidrBlock6s) ([]*net.IPNet, error) {
	sort.Sort(blocks)

	// Coalesce overlapping blocks.
	for i := len(blocks) - 1; i > 0; i-- {
		cmp := blocks[i-1].last
		cmp.Add(cmp, big.NewInt(1))
		if blocks[i].first.Cmp(cmp) <= 0 {
			blocks[i-1].last = blocks[i].last
			if blocks[i].first.Cmp(blocks[i-1].first) < 0 {
				blocks[i-1].first = blocks[i].first
			}
			blocks[i] = nil
		}
	}

	var merged []*net.IPNet
	for _, block := range blocks {
		if block == nil {
			continue
		}

		if err := splitRange6(big.NewInt(0), 0, block.first, block.last, &merged); err != nil {
			return nil, err
		}
	}
	return merged, nil
}
