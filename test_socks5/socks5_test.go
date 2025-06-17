package socks5

import (
	"bytes"
	"testing"
)

func TestSocks5ClientVersionRoundTrip(t *testing.T) {
	original := &Socks5ClientVersion{
		Version:  5,
		NMethods: 3,
		Methods:  []uint8{0, 1, 2},
	}

	// Encode
	encoded, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
	expected := []byte{5, 3, 0, 1, 2}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("MarshalBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseSocks5ClientVersion(encoded)
	if err != nil {
		t.Errorf("ParseSocks5ClientVersion() error = %v", err)
	}

	// Verify round-trip
	if parsed.Version != original.Version || parsed.NMethods != original.NMethods {
		t.Errorf("Round-trip failed: got Version=%d, NMethods=%d", parsed.Version, parsed.NMethods)
	}
	if !bytes.Equal(parsed.Methods, original.Methods) {
		t.Errorf("Methods mismatch: got %v, want %v", parsed.Methods, original.Methods)
	}
}

func TestSocks5ClientRequestIPv4RoundTrip(t *testing.T) {
	original := &Socks5ClientRequest{
		Version:  5,
		Command:  1, // CMD_CONNECT
		Reserved: 0,
		Atype:    1, // ATYPE_IPV4
		Ipv4:     0x08080808, // 8.8.8.8
		DestPort: 80,
	}

	// Encode
	encoded, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
	// Version=5, Command=1, Reserved=0, Atype=1, IPv4=8.8.8.8, Port=80
	expected := []byte{5, 1, 0, 1, 0x08, 0x08, 0x08, 0x08, 0x00, 0x50}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("MarshalBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseSocks5ClientRequest(encoded)
	if err != nil {
		t.Errorf("ParseSocks5ClientRequest() error = %v", err)
	}

	// Verify round-trip
	if parsed.Version != original.Version || parsed.Command != original.Command ||
		parsed.Reserved != original.Reserved || parsed.Atype != original.Atype ||
		parsed.Ipv4 != original.Ipv4 || parsed.DestPort != original.DestPort {
		t.Errorf("IPv4 round-trip failed: got %+v, want %+v", parsed, original)
	}
}

func TestSocks5ClientRequestDomainnameRoundTrip(t *testing.T) {
	domainname := &Domainname{
		Len:  11,
		Name: []byte("example.com"),
	}

	original := &Socks5ClientRequest{
		Version:    5,
		Command:    1, // CMD_CONNECT
		Reserved:   0,
		Atype:      3, // ATYPE_DOMAINNAME
		Domainname: domainname,
		DestPort:   443,
	}

	// Encode
	encoded, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}

	// Decode
	parsed, err := ParseSocks5ClientRequest(encoded)
	if err != nil {
		t.Errorf("ParseSocks5ClientRequest() error = %v", err)
	}

	// Verify round-trip
	if parsed.Version != original.Version || parsed.Command != original.Command ||
		parsed.Reserved != original.Reserved || parsed.Atype != original.Atype ||
		parsed.DestPort != original.DestPort {
		t.Errorf("Domainname round-trip failed: basic fields mismatch")
	}

	if parsed.Domainname.Len != domainname.Len {
		t.Errorf("Domainname length mismatch: got %d, want %d", parsed.Domainname.Len, domainname.Len)
	}
	if !bytes.Equal(parsed.Domainname.Name, domainname.Name) {
		t.Errorf("Domainname name mismatch: got %v, want %v", parsed.Domainname.Name, domainname.Name)
	}
}

func TestSocks5ConstraintValidation(t *testing.T) {
	// Test invalid version
	invalidVersion := &Socks5ClientVersion{
		Version:  4, // Should be 5
		NMethods: 1,
		Methods:  []uint8{0},
	}

	_, err := invalidVersion.MarshalBinary()
	if err == nil {
		t.Error("Expected error for invalid version, got nil")
	}

	// Test invalid command
	invalidCommand := &Socks5ClientRequest{
		Version:  5,
		Command:  99, // Invalid command
		Reserved: 0,
		Atype:    1,
		Ipv4:     0x08080808,
		DestPort: 80,
	}

	_, err = invalidCommand.MarshalBinary()
	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}

	// Test invalid reserved field
	invalidReserved := &Socks5ClientRequest{
		Version:  5,
		Command:  1,
		Reserved: 1, // Should be 0
		Atype:    1,
		Ipv4:     0x08080808,
		DestPort: 80,
	}

	_, err = invalidReserved.MarshalBinary()
	if err == nil {
		t.Error("Expected error for invalid reserved field, got nil")
	}
}

func TestDomainnameValidation(t *testing.T) {
	// Test length mismatch
	invalidDomain := &Domainname{
		Len:  5,
		Name: []byte("example.com"), // 11 bytes, but Len says 5
	}

	_, err := invalidDomain.MarshalBinary()
	if err == nil {
		t.Error("Expected error for domain length mismatch, got nil")
	}
}

func TestComplexSocks5RoundTrip(t *testing.T) {
	// Test multiple round-trips with different address types
	testCases := []struct {
		name    string
		request *Socks5ClientRequest
	}{
		{
			name: "IPv4",
			request: &Socks5ClientRequest{
				Version: 5, Command: 1, Reserved: 0, Atype: 1,
				Ipv4: 0xC0A80001, DestPort: 8080, // 192.168.0.1:8080
			},
		},
		{
			name: "IPv6",
			request: &Socks5ClientRequest{
				Version: 5, Command: 2, Reserved: 0, Atype: 4,
				Ipv6: [16]uint8{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
				DestPort: 443,
			},
		},
		{
			name: "Domain",
			request: &Socks5ClientRequest{
				Version: 5, Command: 3, Reserved: 0, Atype: 3,
				Domainname: &Domainname{Len: 9, Name: []byte("localhost")},
				DestPort:   22,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Multiple encode/decode cycles
			current := tc.request
			for i := 0; i < 3; i++ {
				encoded, err := current.MarshalBinary()
				if err != nil {
					t.Errorf("Round-trip %d MarshalBinary failed: %v", i, err)
				}
				parsed, err := ParseSocks5ClientRequest(encoded)
				if err != nil {
					t.Errorf("Round-trip %d ParseSocks5ClientRequest failed: %v", i, err)
				}
				
				// Basic field verification
				if parsed.Version != current.Version || parsed.Command != current.Command ||
					parsed.Reserved != current.Reserved || parsed.Atype != current.Atype ||
					parsed.DestPort != current.DestPort {
					t.Errorf("Round-trip %d basic fields mismatch", i)
				}
				
				current = parsed
			}
		})
	}
}
