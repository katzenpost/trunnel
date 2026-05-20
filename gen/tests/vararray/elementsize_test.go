package vararray

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseArrayBoundsAccountsForElementSize exercises the tighter form of
// the IDRef bounds check that the trunnel generator now emits for arrays
// of multi-byte element types. The VarArray spec is a u16-length-prefixed
// array of u32 elements, so an honest input of N words must carry at least
// 4*N bytes after the length prefix. Without the element-size division,
// the parser would accept a declared count up to len(remaining) and then
// make() a u32 slice up to 4x larger than the bytes can fill, leaking
// O(buffer) memory per parse. With the division, the bound is exact.
func TestParseArrayBoundsAccountsForElementSize(t *testing.T) {
	// Length prefix says 3 words. To be honest, the buffer needs at
	// least 2 + 4*3 = 14 bytes total. We send 2 + 4*3 - 1 = 13 bytes,
	// which is one byte short of three full u32 elements.
	const declared uint16 = 3
	b := make([]byte, 13)
	binary.BigEndian.PutUint16(b[:2], declared)

	_, err := ParseVarArray(b)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "data too short"),
		"expected 'data too short' error, got: %v", err)

	// A buffer that satisfies the exact size requirement should parse.
	b = make([]byte, 14)
	binary.BigEndian.PutUint16(b[:2], declared)
	if _, err := ParseVarArray(b); err != nil {
		t.Fatalf("baseline parse with exact-sized buffer failed: %v", err)
	}
}
