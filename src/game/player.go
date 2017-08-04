package game

type Player struct {
}

type River struct {
	source int
	target int
}

type Map struct {
	sites  []int
	rivers []River
	mines  []int
}
