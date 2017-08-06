package game

import (
	"math"
	"math/rand"
)

type Random1Player struct {
	BaselinePlayer
	distanceFromOwned [][]int
	totalScore        []int64

	was   []int
	depth []int
	queue []int
}

func (p *Random1Player) MakeMove(moves []Move) Move {
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

	p.was = make([]int, p.NumSites)
	p.depth = make([]int, p.NumSites)
	p.queue = make([]int, p.NumSites)

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

func (p *Random1Player) Name() string { return "random1" }

func (p *Random1Player) getEdgeScore(e Edge) (score int64) {
	if e.Owner >= 0 {
		return
	}

	for i := range p.was {
		p.was[i] = -1
	}

	for mine := range p.Mines {
		rS := p.reachableFromMine[mine][e.Src]
		rD := p.reachableFromMine[mine][e.Dst]

		if rS == rD {
			continue
		}

		if rS {
			d := int64(p.expectedScore(e.Dst, mine))
			score += d * d
		} else {
			d := int64(p.expectedScore(e.Src, mine))
			score += d * d
		}
	}

	return
}

func (p *Random1Player) countBonus(u, v, mine int) (bonus int64) {
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

func (p *Random1Player) calcDegreesOfFreedom(u int) (d int) {
	for _, e := range p.Edges[u] {
		edge := &p.AllEdges[e]
		if edge.Owner < 0 {
			d++
		}
	}
	return
}

func (p *Random1Player) expectedScore(u, mine int) (score float64) {
	const depthLimit = 10
	const discount = 0.95

	qh, qt := 0, 0

	p.was[u] = mine
	p.queue[qt] = u
	p.depth[u] = 0
	qt++

	for qh < qt {
		u := p.queue[qh]
		qh++

		score += math.Pow(discount, float64(p.depth[u])) * float64(p.Distance[mine][u]) * float64(p.Distance[mine][u])
		if p.depth[u] == depthLimit {
			continue
		}
		for _, e := range p.Edges[u] {
			edge := p.AllEdges[e]
			if edge.Owner < 0 && p.was[edge.Dst] != mine {
				v := edge.Dst
				p.was[v] = mine
				p.queue[qt] = v
				qt++
				p.depth[v] = p.depth[u] + 1
			}
		}
	}

	return
}

func (p *Random1Player) fromMine(e Edge) bool {
	for _, m := range p.Mines {
		if e.Src == m {
			return true
		}
	}
	return false
}
