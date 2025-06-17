package arrays

import (
	"bytes"
	"testing"
)

func TestSimpleArrayRoundTrip(t *testing.T) {
	original := &SimpleArray{
		Count: 3,
		Data:  []uint8{10, 20, 30},
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{3, 10, 20, 30}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Test MarshalBinary method
	marshaled, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
	if !bytes.Equal(marshaled, expected) {
		t.Errorf("MarshalBinary() = %v, want %v", marshaled, expected)
	}

	// Decode
	parsed, err := ParseSimpleArray(encoded)
	if err != nil {
		t.Errorf("ParseSimpleArray() error = %v", err)
	}

	// Verify round-trip
	if parsed.Count != original.Count {
		t.Errorf("Count mismatch: got %d, want %d", parsed.Count, original.Count)
	}
	if !bytes.Equal(parsed.Data, original.Data) {
		t.Errorf("Data mismatch: got %v, want %v", parsed.Data, original.Data)
	}
}

func TestFixedArrayRoundTrip(t *testing.T) {
	original := &FixedArray{
		Values:  [4]uint8{1, 2, 3, 4},
		Numbers: [3]uint16{256, 512, 1024},
	}

	// Encode
	encoded := original.encodeBinary()
	// Values: 1, 2, 3, 4
	// Numbers: 256 (0x0100), 512 (0x0200), 1024 (0x0400)
	expected := []byte{1, 2, 3, 4, 0x01, 0x00, 0x02, 0x00, 0x04, 0x00}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Test MarshalBinary method
	marshaled, err := original.MarshalBinary()
	if err != nil {
		t.Errorf("MarshalBinary() error = %v", err)
	}
	if !bytes.Equal(marshaled, expected) {
		t.Errorf("MarshalBinary() = %v, want %v", marshaled, expected)
	}

	// Decode
	parsed, err := ParseFixedArray(encoded)
	if err != nil {
		t.Errorf("ParseFixedArray() error = %v", err)
	}

	// Verify round-trip
	if parsed.Values != original.Values {
		t.Errorf("Values mismatch: got %v, want %v", parsed.Values, original.Values)
	}
	if parsed.Numbers != original.Numbers {
		t.Errorf("Numbers mismatch: got %v, want %v", parsed.Numbers, original.Numbers)
	}
}

func TestSimpleArrayValidation(t *testing.T) {
	// Test count/data length mismatch
	invalid := &SimpleArray{
		Count: 5,
		Data:  []uint8{1, 2, 3}, // Only 3 elements, but count says 5
	}

	_, err := invalid.MarshalBinary()
	if err == nil {
		t.Error("Expected error for count/data length mismatch, got nil")
	}
}

func TestEmptyArray(t *testing.T) {
	original := &SimpleArray{
		Count: 0,
		Data:  []uint8{},
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{0}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseSimpleArray(encoded)
	if err != nil {
		t.Errorf("ParseSimpleArray() error = %v", err)
	}

	// Verify round-trip
	if parsed.Count != 0 || len(parsed.Data) != 0 {
		t.Errorf("Empty array round-trip failed: got Count=%d, Data=%v", parsed.Count, parsed.Data)
	}
}
