package utils

import (
	"fmt"
	"net"
	"strings"
)

func CidrZeroBased(cidr string) (string, error) {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	return cidrNet.String(), nil
}

func CidrOneBased(cidr string) (string, error) {
	_, cidrNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	ip4 := cidrNet.IP.To4()
	if ip4 == nil {
		return "", fmt.Errorf("CidrOneBased only supports IPv4, got %q", cidr)
	}
	ip4[3]++
	cidrNet.IP = ip4

	return cidrNet.String(), nil
}

// IsIPv4 checks if the provided address is a valid IPv4 address.
func IsIPv4(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 checks if the provided address is a valid IPv6 address.
func IsIPv6(address string) bool {
	// Handle zone index if present
	if idx := strings.Index(address, "%"); idx != -1 {
		address = address[:idx]
	}

	// Handle IPv4-mapped addresses
	isIPv4Mapped := strings.Contains(address, "::ffff:") && strings.Count(address, ".") == 3

	ip := net.ParseIP(address)
	if ip == nil || (!isIPv4Mapped && ip.To4() != nil) {
		return false
	}
	return true
}
