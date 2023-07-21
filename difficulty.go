package main

import "math/rand"

var (
	difficultyE = [3]int{12, 6, 2}
	difficultyN = [3]int{6, 8, 6}
	difficultyL = [3]int{2, 6, 12}
)

func difficultyRandom() [3]int {
	r := rand.Intn(30)
	n := 8
	if r < 11 {
		n = 6
	} else if r < 21 {
		n = 7
	}
	e := 2 + rand.Intn(17-n)
	return [3]int{e, n, 20 - e - n}
}
