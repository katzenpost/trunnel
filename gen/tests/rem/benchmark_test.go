package rem

import "testing"

func BenchmarkParseByteTail(b *testing.B) {
	data := make([]byte, 4+64*1024)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := ParseRem(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalByteTail(b *testing.B) {
	message := &Rem{Tail: make([]byte, 64*1024)}
	b.SetBytes(int64(4 + len(message.Tail)))
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := message.MarshalBinary(); err != nil {
			b.Fatal(err)
		}
	}
}
