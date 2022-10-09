package arrays

import "math/rand"

func Contains[T comparable](arr []T, e T) bool {
	for _, e1 := range arr {
		if e1 == e {
			return true
		}
	}
	return false
}

func ShuffleN[T any](rand *rand.Rand, arr []T, n int) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}
	for i := 0; i < n; i++ {
		j := i + rand.Intn(n-i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func Any[T any](arr []T, f func(e T) bool) bool {
	for _, e := range arr {
		if !f(e) {
			return true
		}
	}
	return false
}
