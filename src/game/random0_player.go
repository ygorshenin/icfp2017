package game

import (
	"math/rand"
)

type Random0Player struct {
	BaselinePlayer
	distanceFromOwned [][]int
	totalScore        []int64
}

func (p *Random0Player) MakeMove(moves []Move) Move {
	p.BaselinePlayer.PrepareForMove(moves)

	for _, e := range p.AllEdges {
		if e.Owner < 0 && p.fromMine(e) {
			return p.MakeClaimMove(e.Src, e.Dst)
		}
	}

	p.distanceFromOwned = make([][]int, len(p.Mines))
	for i := 0; i < len(p.Mines); i++ {
		was := make([]bool, p.NumSites)
		copy(was, p.reachableFromMine[i])
		p.distanceFromOwned[i] = p.MSSP(was)
	}

	scores := make([]int64, len(p.AllEdges))
	var bestScore int64
	for i, e := range p.AllEdges {
		scores[i] = p.getEdgeScore(e)
		if bestScore < scores[i] {
			bestScore = scores[i]
		}
	}

	if bestScore == 0 {
		return p.MakePassMove()
	}

	r := rand.New(rand.NewSource(42))
	visited := 0
	var move Move
	for i, score := range scores {
		if score < bestScore {
			continue
		}
		visited++
		if r.Intn(visited) == 0 {
			e := &p.AllEdges[i]
			move = p.MakeClaimMove(e.Src, e.Dst)
		}
	}

	return move
}

func (p *Random0Player) Name() string { return "random0" }

func (p *Random0Player) getEdgeScore(e Edge) (score int64) {
	if e.Owner >= 0 {
		return
	}

	was := make([]int, p.NumSites)
	for i := range was {
		was[i] = -1
	}

	for mine := range p.Mines {
		rS := p.reachableFromMine[mine][e.Src]
		rD := p.reachableFromMine[mine][e.Dst]

		if rS == rD {
			continue
		}

		if rS {
			d := p.expectedScore(e.Dst, mine, 0, 10, was)
			score += d * d
		} else {
			d := p.expectedScore(e.Src, mine, 0, 10, was)
			score += d * d
		}
	}

	return
}

func (p *Random0Player) countBonus(u, v, mine int) (bonus int64) {
	for _, e := range p.Edges[u] {
		edge := &p.AllEdges[e]
		if edge.Owner >= 0 {
			continue
		}
		if edge.Dst == v || p.reachableFromMine[mine][edge.Dst] {
			continue
		}
		d := int64(p.Distance[mine][edge.Dst])
		bonus += d * d
	}
	return bonus
}

func (p *Random0Player) calcDegreesOfFreedom(u int) (d int) {
	for _, e := range p.Edges[u] {
		edge := &p.AllEdges[e]
		if edge.Owner < 0 {
			d++
		}
	}
	return
}

func (p *Random0Player) expectedScore(u, mine, depth, limit int, was []int) (score int64) {
	was[u] = mine
	score += int64(p.Distance[mine][u]) * int64(p.Distance[mine][u])
	if depth == limit {
		return
	}
	for _, e := range p.Edges[u] {
		edge := p.AllEdges[e]
		if edge.Owner < 0 && was[edge.Dst] != mine {
			score += p.expectedScore(edge.Dst, mine, depth+1, limit, was)
		}
	}
	return
}

func (p *Random0Player) fromMine(e Edge) bool {
	for _, m := range p.Mines {
		if e.Src == m {
			return true
		}
	}
	return false
}
