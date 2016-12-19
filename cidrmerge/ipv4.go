package cidrmerge

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ipv4ToUInt32 converts an IPv4 address to an unsigned 32-bit integer.
func ipv4ToUInt32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip)
}

// uint32ToIPV4 converts an unsigned 32-bit integer to an IPv4 address.
func uint32ToIPV4(addr uint32) net.IP {
	ip := make([]byte, net.IPv4len)
	binary.BigEndian.PutUint32(ip, addr)
	return ip
}

// The following functions are inspired by http://www.cs.colostate.edu/~somlo/iprange.c.

// setBit sets the specified bit in an address to 0 or 1.
func setBit4(addr uint32, bit uint, val int) uint32 {
	if val == 0 {
		return addr & ^(1 << (32 - bit))
	}
	return addr | (1 << (32 - bit))
}

// netmask returns the netmask for the specified prefix.
func netmask4(prefix uint) uint32 {
	if prefix == 0 {
		return 0
	}
	return ^uint32((1 << (32 - prefix)) - 1)
}

// broadcast returns the broadcast address for the given address and prefix.
func broadcast4(addr uint32, prefix uint) uint32 {
	return addr | ^netmask4(prefix)
}

// network returns the network address for the given address and prefix.
func network4(addr uint32, prefix uint) uint32 {
	return addr & netmask4(prefix)
}

// splitRange recursively computes the CIDR blocks to cover the range lo to hi.
func splitRange4(addr uint32, prefix uint, lo, hi uint32, cidrs *[]*net.IPNet) error {
	if prefix > 32 {
		return fmt.Errorf("Invalid mask size: %d", prefix)
	}

	bc := broadcast4(addr, prefix)
	if (lo < addr) || (hi > bc) {
		return fmt.Errorf("%d, %d out of range for network %d/%d, broadcast %d", lo, hi, addr, prefix, bc)
	}

	if (lo == addr) && (hi == bc) {
		cidr := net.IPNet{IP: uint32ToIPV4(addr), Mask: net.CIDRMask(int(prefix), 8*net.IPv4len)}
		*cidrs = append(*cidrs, &cidr)
		return nil
	}

	prefix++
	lowerHalf := addr
	upperHalf := setBit4(addr, prefix, 1)
	if hi < upperHalf {
		return splitRange4(lowerHalf, prefix, lo, hi, cidrs)
	} else if lo >= upperHalf {
		return splitRange4(upperHalf, prefix, lo, hi, cidrs)
	} else {
		err := splitRange4(lowerHalf, prefix, lo, broadcast4(lowerHalf, prefix), cidrs)
		if err != nil {
			return err
		}
		return splitRange4(upperHalf, prefix, upperHalf, hi, cidrs)
	}
}
