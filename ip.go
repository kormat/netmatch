package netmatch

import "net"

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// Key converts a net.IP to a [16]byte.
// If ip is an IPv4 address, it will be prefixed with the net.v4InV6Prefix
// to get a valid IPv6 address.
// The resulting array can then be used for the Trie.
func Key(ip net.IP) [16]byte {
	var array [16]byte

	if ip4 := ip.To4(); ip4 != nil {
		copy(array[:], v4InV6Prefix)
		copy(array[12:], ip4)
	} else {
		copy(array[:], ip)
	}
	return array
}

// ParseNetwork parses a network in CIDR notation and returns everything
// necessary to add that network to the Trie.
func ParseNetwork(network string) ([16]byte, int, error) {
	_, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return [16]byte{}, 0, err
	}

	key := Key(ipnet.IP)
	size, _ := ipnet.Mask.Size()
	if ipnet.IP.To4() != nil {
		size = size + 12*8
	}

	return key, size, nil
}
