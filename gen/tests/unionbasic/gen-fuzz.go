// Code generated by trunnel. DO NOT EDIT.

//go:build gofuzz
// +build gofuzz

package unionbasic

func FuzzDate(data []byte) int {
	_, err := ParseDate(data)
	if err != nil {
		return 0
	}
	return 1
}

func FuzzBasic(data []byte) int {
	_, err := ParseBasic(data)
	if err != nil {
		return 0
	}
	return 1
}
