package random

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVar(t *testing.T) {
	assert.Equal(t, Digits, "0123456789")
	assert.Equal(t, UppercaseAlpha, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	assert.Equal(t, LowercaseAlpha, "abcdefghijklmnopqrstuvwxyz")
	assert.Equal(t, VisibleASCII, `!"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`+"`abcdefghijklmnopqrstuvwxyz{|}~")
}

func assertVocab(t *testing.T, s string, vocab ...string) {
	combined := []byte(strings.Join(vocab, ""))
	n := len(combined)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i] < combined[j]
	})
	for _, b := range []byte(s) {
		j := sort.Search(n, func(i int) bool {
			return combined[i] >= b
		})
		assert.Less(t, j, n)
	}
}

func TestString(t *testing.T) {
	s := String(100, VisibleASCII)
	assertVocab(t, s, VisibleASCII)

	s = AlphaString(100)
	assertVocab(t, s, UppercaseAlpha, LowercaseAlpha)

	s = AlphaNumericString(100)
	assertVocab(t, s, UppercaseAlpha, LowercaseAlpha, Digits)
}

func TestNumbers(t *testing.T) {
	assert.NotEmpty(t, Int32())
	assert.NotEmpty(t, Int64())
	assert.NotEmpty(t, Float32())
	assert.NotEmpty(t, Float64())
	min := 10
	max := 999
	n := Between(min, max)
	assert.LessOrEqual(t, n, max)
	assert.GreaterOrEqual(t, n, min)
	Bool()
}
