package utils

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var hexLetters = []rune("0123456789abcdef")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func RandHexString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(hexLetters))]
	}
	return string(b)
}

func RandIndex(size int) []int {
	result := make([]int, size)
	for i := range result {
		result[i] = i
	}
	rand.Shuffle(size, func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}
