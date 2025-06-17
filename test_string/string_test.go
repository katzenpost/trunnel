package strings

import (
	"bytes"
	"testing"
)

func TestStringRoundTrip(t *testing.T) {
	original := &Message{
		Type:    1,
		Text:    "Hello, World!",
		Padding: 42,
	}

	// Encode
	encoded := original.encodeBinary()
	// Type: 1, Text: "Hello, World!" + null terminator, Padding: 42
	expected := []byte{1, 'H', 'e', 'l', 'l', 'o', ',', ' ', 'W', 'o', 'r', 'l', 'd', '!', 0, 42}
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
	parsed, err := ParseMessage(encoded)
	if err != nil {
		t.Errorf("ParseMessage() error = %v", err)
	}

	// Verify round-trip
	if parsed.Type != original.Type {
		t.Errorf("Type mismatch: got %d, want %d", parsed.Type, original.Type)
	}
	if parsed.Text != original.Text {
		t.Errorf("Text mismatch: got %q, want %q", parsed.Text, original.Text)
	}
	if parsed.Padding != original.Padding {
		t.Errorf("Padding mismatch: got %d, want %d", parsed.Padding, original.Padding)
	}
}

func TestEmptyStringRoundTrip(t *testing.T) {
	original := &Message{
		Type:    2,
		Text:    "",
		Padding: 255,
	}

	// Encode
	encoded := original.encodeBinary()
	expected := []byte{2, 0, 255} // Type, empty string (just null terminator), Padding
	if !bytes.Equal(encoded, expected) {
		t.Errorf("encodeBinary() = %v, want %v", encoded, expected)
	}

	// Decode
	parsed, err := ParseMessage(encoded)
	if err != nil {
		t.Errorf("ParseMessage() error = %v", err)
	}

	// Verify round-trip
	if parsed.Text != "" {
		t.Errorf("Empty string round-trip failed: got %q, want %q", parsed.Text, "")
	}
}

func TestUnicodeStringRoundTrip(t *testing.T) {
	original := &Message{
		Type:    3,
		Text:    "Hello, 世界! 🌍",
		Padding: 100,
	}

	// Encode
	encoded := original.encodeBinary()

	// Decode
	parsed, err := ParseMessage(encoded)
	if err != nil {
		t.Errorf("ParseMessage() error = %v", err)
	}

	// Verify round-trip
	if parsed.Text != original.Text {
		t.Errorf("Unicode string round-trip failed: got %q, want %q", parsed.Text, original.Text)
	}
}

func TestStringWithSpecialChars(t *testing.T) {
	// Test string with various special characters (but no null bytes)
	original := &Message{
		Type:    4,
		Text:    "Line1\nLine2\tTabbed\rCarriage",
		Padding: 0,
	}

	// Encode
	encoded := original.encodeBinary()

	// Decode
	parsed, err := ParseMessage(encoded)
	if err != nil {
		t.Errorf("ParseMessage() error = %v", err)
	}

	// Verify round-trip
	if parsed.Text != original.Text {
		t.Errorf("Special chars string round-trip failed: got %q, want %q", parsed.Text, original.Text)
	}
}
