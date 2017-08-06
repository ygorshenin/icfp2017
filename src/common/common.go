package common

import (
	"game"
)

type CompressedIndex struct {
	Forward  map[int]int `json:"forward"`
	Backward []int       `json:"backward"`
}

func (index *CompressedIndex) Setup(values []int) {
	n := 0
	index.Forward = make(map[int]int)
	for _, v := range values {
		_, ok := index.Forward[v]
		if !ok {
			index.Forward[v] = n
			n++
		}
	}
	index.Backward = make([]int, n)
	for k, v := range index.Forward {
		index.Backward[v] = k
	}
}

type Site struct {
	Id int      `json:"id"`
	X  *float64 `json:"x,omitempty"`
	Y  *float64 `json:"y,omitempty"`
}

// The Map structure from server is inconvinient, this is a wrapper.
type Map struct {
	Sites  []Site       `json:"sites"`
	Rivers []game.River `json:"rivers"`
	Mines  []int        `json:"mines"`
}
