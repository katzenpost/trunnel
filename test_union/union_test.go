package unions

import (
	"bytes"
	"testing"
)

func TestUnionByteDataRoundTrip(t *testing.T) {
	original := &Packet{
		Type:     1,
		ByteData: 255,
		// Other fields should be ignored when Type=1
		WordData:   999,
		StringData: "ignored",
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{1, 255} // Type=1, ByteData=255
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParsePacket(encoded)
	if err != nil {
		t.Errorf("ParsePacket() error = %v", err)
	}

	// Verify round-trip
	if parsed.Type != 1 || parsed.ByteData != 255 {
		t.Errorf("Byte data round-trip failed: got Type=%d, ByteData=%d", parsed.Type, parsed.ByteData)
	}
}

func TestUnionWordDataRoundTrip(t *testing.T) {
	original := &Packet{
		Type:     2,
		WordData: 0x1234,
		// Other fields should be ignored when Type=2
		ByteData:   99,
		StringData: "ignored",
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{2, 0x12, 0x34} // Type=2, WordData=0x1234 (big-endian)
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParsePacket(encoded)
	if err != nil {
		t.Errorf("ParsePacket() error = %v", err)
	}

	// Verify round-trip
	if parsed.Type != 2 || parsed.WordData != 0x1234 {
		t.Errorf("Word data round-trip failed: got Type=%d, WordData=%d", parsed.Type, parsed.WordData)
	}
}

func TestUnionStringDataRoundTrip(t *testing.T) {
	original := &Packet{
		Type:       3,
		StringData: "Hello Union!",
		// Other fields should be ignored when Type=3
		ByteData: 99,
		WordData: 999,
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{3, 'H', 'e', 'l', 'l', 'o', ' ', 'U', 'n', 'i', 'o', 'n', '!', 0}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParsePacket(encoded)
	if err != nil {
		t.Errorf("ParsePacket() error = %v", err)
	}

	// Verify round-trip
	if parsed.Type != 3 || parsed.StringData != "Hello Union!" {
		t.Errorf("String data round-trip failed: got Type=%d, StringData=%q", parsed.Type, parsed.StringData)
	}
}

func TestUnionEmptyString(t *testing.T) {
	original := &Packet{
		Type:       3,
		StringData: "",
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{3, 0} // Type=3, empty string (just null terminator)
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParsePacket(encoded)
	if err != nil {
		t.Errorf("ParsePacket() error = %v", err)
	}

	// Verify round-trip
	if parsed.StringData != "" {
		t.Errorf("Empty string union round-trip failed: got %q", parsed.StringData)
	}
}

func TestUnionInvalidType(t *testing.T) {
	// Test parsing with invalid type (should fail due to "default: fail")
	invalidData := []byte{4, 123} // Type=4 is not allowed

	_, err := ParsePacket(invalidData)
	if err == nil {
		t.Error("Expected error for invalid union type, got nil")
	}
}

func TestUnionMultipleRoundTrips(t *testing.T) {
	testCases := []struct {
		name string
		packet *Packet
	}{
		{
			name: "ByteData",
			packet: &Packet{Type: 1, ByteData: 42},
		},
		{
			name: "WordData",
			packet: &Packet{Type: 2, WordData: 0xABCD},
		},
		{
			name: "StringData",
			packet: &Packet{Type: 3, StringData: "Test String"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Multiple encode/decode cycles
			current := tc.packet
			for i := 0; i < 5; i++ {
				encoded := current.encodeBinary()
				parsed, err := ParsePacket(encoded)
				if err != nil {
					t.Errorf("Round-trip %d failed: %v", i, err)
				}
				
				// Verify the active field based on type
				switch current.Type {
				case 1:
					if parsed.ByteData != current.ByteData {
						t.Errorf("Round-trip %d ByteData mismatch: got %d, want %d", i, parsed.ByteData, current.ByteData)
					}
				case 2:
					if parsed.WordData != current.WordData {
						t.Errorf("Round-trip %d WordData mismatch: got %d, want %d", i, parsed.WordData, current.WordData)
					}
				case 3:
					if parsed.StringData != current.StringData {
						t.Errorf("Round-trip %d StringData mismatch: got %q, want %q", i, parsed.StringData, current.StringData)
					}
				}
				current = parsed
			}
		})
	}
}
