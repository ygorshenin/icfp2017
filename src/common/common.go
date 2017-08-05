package common

import "game"

type Site struct {
	Id int `json:"id"`
}

// The Map structure from server is inconvinient, this is a wrapper.
type Map struct {
	Sites  []Site       `json:"sites"`
	Rivers []game.River `json:"rivers"`
	Mines  []int        `json:"mines"`
}

func ToGameMap(m *Map) game.Map {
	sites := make([]int, len(m.Sites), len(m.Sites))
	for i, site := range m.Sites {
		sites[i] = site.Id
	}
	return game.Map{Sites: sites, Rivers: m.Rivers, Mines: m.Mines}
}
