package bytesfield

import (
	"bytes"
	"testing"
)

func TestBytesFieldRoundTrip(t *testing.T) {
	original := &IRecv{
		Bytes: [8]uint8{1, 2, 3, 4, 5, 6, 7, 8},
	}

	// Encode using MarshalBinary
	encoded, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("MarshalBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseIRecv(encoded)
	if err != nil {
		t.Errorf("ParseIRecv() error = %v", err)
	}

	// Verify round-trip
	if parsed.Bytes != original.Bytes {
		t.Errorf("Round-trip failed: got %+v, want %+v", parsed.Bytes, original.Bytes)
	}
}

func TestBytesFieldAccess(t *testing.T) {
	// Test that we can access the Bytes field without conflict
	recv := &IRecv{}
	
	// Set the Bytes field
	recv.Bytes[0] = 42
	recv.Bytes[7] = 99
	
	// Verify we can access it
	if recv.Bytes[0] != 42 || recv.Bytes[7] != 99 {
		t.Errorf("Field access failed: got Bytes[0]=%d, Bytes[7]=%d", recv.Bytes[0], recv.Bytes[7])
	}
	
	// Verify we can call MarshalBinary
	_, err := recv.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
}
