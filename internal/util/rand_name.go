package util

import "math/rand/v2"

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandName Generates a random, 32-character, Docker-compatible container name.
func RandName() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}

	return string(b)
}
