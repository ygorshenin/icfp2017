package game

import (
	"math"
	"math/rand"
)

type Random1Player struct {
	BaselinePlayer
	distanceFromOwned [][]int
	totalScore        []int64

	numFreeEdges int
	was          []int64
	depth        []int
	queue        []int
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

	p.was = make([]int64, p.NumSites)
	p.depth = make([]int, p.NumSites)
	p.queue = make([]int, p.NumSites)

	p.numFreeEdges = 0
	for _, e := range p.AllEdges {
		if e.Owner < 0 {
			p.numFreeEdges++
		}
	}

	r := rand.New(rand.NewSource(42))
	move := p.MakePassMove()
	var bestScore int64
	var visited int
	for _, e := range p.AllEdges {
		score := p.getEdgeScore(e)
		if score == 0 {
			continue
		}

		if bestScore < score {
			bestScore = score
			visited = 1
			move = p.MakeClaimMove(e.Src, e.Dst)
			continue
		}

		if score == bestScore {
			visited++
			if r.Intn(visited) == 0 {
				move = p.MakeClaimMove(e.Src, e.Dst)
			}
		}
	}

	return move
}

func (p *Random1Player) Name() string { return "random1" }

func (p *Random1Player) getEdgeScore(e Edge) (score int64) {
	if e.Owner >= 0 {
		return
	}

	for mine := range p.Mines {
		rS := p.reachableFromMine[mine][e.Src]
		rD := p.reachableFromMine[mine][e.Dst]

		if rS == rD {
			continue
		}

		var d int64

		if rS {
			d = int64(p.expectedScore(e.Dst, mine, e.Id))
		} else {
			d = int64(p.expectedScore(e.Src, mine, e.Id))
		}

		score += d * d
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

func (p *Random1Player) expectedScore(u, mine, edgeId int) (score float64) {
	const depthLimit = 20
	edgesLimit := 10000000 / len(p.Mines) / p.numFreeEdges
	discount := 0.95
	mark := int64(len(p.AllEdges))*int64(mine) + int64(edgeId)

	qh, qt := 0, 0

	p.was[u] = mark
	p.queue[qt] = u
	p.depth[u] = 0
	qt++

	visitedEdges := 0
	for qh < qt && visitedEdges < edgesLimit {
		u := p.queue[qh]
		qh++

		score += math.Pow(discount, float64(p.depth[u])) * float64(p.Distance[mine][u]) * float64(p.Distance[mine][u])
		if p.depth[u] == depthLimit {
			continue
		}
		for _, e := range p.Edges[u] {
			visitedEdges++
			edge := p.AllEdges[e]
			if edge.Owner < 0 && p.was[edge.Dst] != mark {
				v := edge.Dst
				p.was[v] = mark
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
