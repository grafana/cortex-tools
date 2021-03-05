// Package to generate random strings
package randomstring

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

const (
	// [a-z]+
	CharsetAlphaLow = "abcdefghijklmnopqrstuvwxyz"
	// [A-Z]+
	CharsetAlphaUp = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// [0-9]+
	CharsetNum = "0123456789"
)

var (
	rsg *Generator
)

func init() {
	rsg, _ = NewGenerator(CharsetAlphaLow, CharsetAlphaUp, CharsetNum)
}

// Generate a random string of length `n` using the default generator
func Generate(n int) string {
	return rsg.Generate(n)
}

// A random string generator
type Generator struct {
	charset       string
	charsetLength int
	letterIdxBits uint
}

// Create a new generator with the given charset
func NewGenerator(charsets ...string) (*Generator, error) {
	ret := &Generator{}
	return ret.WithCharsets(charsets...)
}

// Change the charset of the generator
func (rsg *Generator) WithCharsets(cs ...string) (*Generator, error) {
	rsg.charset = ""
	for _, c := range cs {
		rsg.charset += c
	}
	rsg.charsetLength = len(rsg.charset)
	letterIdxBits := math.Ceil(math.Log2(float64(rsg.charsetLength)))
	if letterIdxBits == 0 {
		return nil, errors.New("charset too long")
	}
	rsg.letterIdxBits = uint(letterIdxBits)
	return rsg, nil
}

// Generate a random string of length `n`
func (rsg *Generator) Generate(n int) string {
	var letterIdxMask int64 = 1<<rsg.letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	var letterIdxMax = 63 / rsg.letterIdxBits          // # of letter indices fitting in 63 bits
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < rsg.charsetLength {
			b[i] = rsg.charset[idx]
			i--
		}
		cache >>= rsg.letterIdxBits
		remain--
	}

	return string(b)
}
