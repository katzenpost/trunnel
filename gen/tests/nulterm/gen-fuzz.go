// Code generated by trunnel. DO NOT EDIT.

//go:build gofuzz
// +build gofuzz

package nulterm

func FuzzNulTerm(data []byte) int {
	_, err := ParseNulTerm(data)
	if err != nil {
		return 0
	}
	return 1
}
