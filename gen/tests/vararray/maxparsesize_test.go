package vararray

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMaxParseSizeRejectsOversizedInput exercises the package-level
// MaxParseSize cap that the trunnel generator emits ahead of every Parse...
// convenience constructor. The cap shields the parser from an upstream that
// hands it an unbounded buffer (for example, a wire layer whose own cap is
// generous, or a caller who forgets to enforce one). Without the cap, a
// well-formed message that happens to be very large would still allocate
// proportionally large memory; with the cap, the constructor rejects the
// input outright before any work begins.
//
// The struct method Parse(data, ...) remains uncapped on purpose so that
// recursive struct refs and in-stream parses do not re-validate on every
// step.
func TestMaxParseSizeRejectsOversizedInput(t *testing.T) {
	prev := MaxParseSize
	t.Cleanup(func() { MaxParseSize = prev })

	// Build a well-formed VarArray of one u32 element, plus an honest
	// length prefix: 2 bytes of u16 length + 4 bytes of u32 payload.
	b := make([]byte, 6)
	binary.BigEndian.PutUint16(b[:2], 1)
	binary.BigEndian.PutUint32(b[2:], 0x01020304)

	// With the default cap, the input parses cleanly.
	if _, err := ParseVarArray(b); err != nil {
		t.Fatalf("baseline parse should succeed, got: %v", err)
	}

	// Tighten the cap to one byte below the input length and confirm
	// the constructor rejects with the expected error.
	MaxParseSize = len(b) - 1
	_, err := ParseVarArray(b)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "MaxParseSize"),
		"expected MaxParseSize error, got: %v", err)

	// The struct's own Parse method does not honour the cap; the cap
	// applies only to the public Parse... convenience constructor.
	v := new(VarArray)
	if _, err := v.Parse(b); err != nil {
		t.Fatalf("direct Parse should not consult MaxParseSize, got: %v", err)
	}
}
