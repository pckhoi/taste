package random

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"strings"
)

func asciiRange(min, max uint8) string {
	n := max - min + 1
	result := make([]byte, n)
	for i := uint8(0); i < n; i++ {
		result[i] = min + uint8(i)
	}
	return string(result)
}

var (
	Digits         = asciiRange(48, 57)
	UppercaseAlpha = asciiRange(65, 90)
	LowercaseAlpha = asciiRange(97, 122)
	VisibleASCII   = asciiRange(33, 126)
)

func String(n int, vocab ...string) string {
	combined := strings.Join(vocab, "")
	m := big.NewInt(int64(len(combined)))
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		bi, err := rand.Int(rand.Reader, m)
		if err != nil {
			panic(err)
		}
		s[i] = combined[bi.Int64()]
	}
	return string(s)
}

func AlphaString(n int) string {
	return String(n, UppercaseAlpha, LowercaseAlpha)
}

func AlphaNumericString(n int) string {
	return String(n, Digits, UppercaseAlpha, LowercaseAlpha)
}

func Int32() int32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	var num int32
	if err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &num); err != nil {
		panic(err)
	}
	return num
}

func Between(min, max int) int {
	bi, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		panic(err)
	}
	return min + int(bi.Int64())
}

func Int64() int64 {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	i, err := binary.ReadVarint(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	return i
}

func Float32() float32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	var num float32
	if err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &num); err != nil {
		panic(err)
	}
	return num
}

func Float64() float64 {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	var num float64
	if err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &num); err != nil {
		panic(err)
	}
	return num
}

func Bool() bool {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b[0] > 127
}
