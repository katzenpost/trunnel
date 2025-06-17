package nested

import (
	"bytes"
	"testing"
)

func TestNestedStructRoundTrip(t *testing.T) {
	original := &Outer{
		Tag: 42,
		Data: &Inner{
			Value:  10,
			Number: 256,
		},
		Count: 2,
		Items: []*Inner{
			{Value: 20, Number: 512},
			{Value: 30, Number: 1024},
		},
	}

	// Encode
	encoded := original.encodeBinary()
	// Tag: 42
	// Data: Value=10, Number=256 (0x0100)
	// Count: 2
	// Items[0]: Value=20, Number=512 (0x0200)
	// Items[1]: Value=30, Number=1024 (0x0400)
	expected := []byte{42, 10, 0x01, 0x00, 2, 20, 0x02, 0x00, 30, 0x04, 0x00}
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
	parsed, err := ParseOuter(encoded)
	if err != nil {
		t.Errorf("ParseOuter() error = %v", err)
	}

	// Verify round-trip
	if parsed.Tag != original.Tag {
		t.Errorf("Tag mismatch: got %d, want %d", parsed.Tag, original.Tag)
	}
	if parsed.Data.Value != original.Data.Value || parsed.Data.Number != original.Data.Number {
		t.Errorf("Data mismatch: got %+v, want %+v", parsed.Data, original.Data)
	}
	if parsed.Count != original.Count {
		t.Errorf("Count mismatch: got %d, want %d", parsed.Count, original.Count)
	}
	if len(parsed.Items) != len(original.Items) {
		t.Errorf("Items length mismatch: got %d, want %d", len(parsed.Items), len(original.Items))
	}
	for i, item := range parsed.Items {
		if item.Value != original.Items[i].Value || item.Number != original.Items[i].Number {
			t.Errorf("Items[%d] mismatch: got %+v, want %+v", i, item, original.Items[i])
		}
	}
}

func TestInnerStructRoundTrip(t *testing.T) {
	original := &Inner{
		Value:  255,
		Number: 65535,
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{255, 0xFF, 0xFF}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseInner(encoded)
	if err != nil {
		t.Errorf("ParseInner() error = %v", err)
	}

	// Verify round-trip
	if parsed.Value != original.Value || parsed.Number != original.Number {
		t.Errorf("Round-trip failed: got %+v, want %+v", parsed, original)
	}
}

func TestEmptyNestedArray(t *testing.T) {
	original := &Outer{
		Tag: 1,
		Data: &Inner{
			Value:  5,
			Number: 100,
		},
		Count: 0,
		Items: []*Inner{},
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{1, 5, 0x00, 0x64, 0}
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseOuter(encoded)
	if err != nil {
		t.Errorf("ParseOuter() error = %v", err)
	}

	// Verify round-trip
	if parsed.Count != 0 || len(parsed.Items) != 0 {
		t.Errorf("Empty nested array round-trip failed: got Count=%d, Items=%v", parsed.Count, parsed.Items)
	}
}

func TestNestedValidation(t *testing.T) {
	// Test count/items length mismatch
	invalid := &Outer{
		Tag: 1,
		Data: &Inner{Value: 1, Number: 1},
		Count: 3,
		Items: []*Inner{{Value: 1, Number: 1}}, // Only 1 item, but count says 3
	}

	_, err := invalid.MarshalBinary()
	if err == nil {
		t.Error("Expected error for count/items length mismatch, got nil")
	}
}
