package game

import (
	"math/rand"
)

type Random0Player struct {
	BaselinePlayer
	distanceFromOwned [][]int
}

func (p *Random0Player) MakeMove(moves []Move) Move {
	p.BaselinePlayer.PrepareForMove(moves)
	p.distanceFromOwned = make([][]int, len(p.Mines))
	for i := 0; i < len(p.Mines); i++ {
		was := make([]bool, p.NumSites)
		copy(was, p.reachableFromMine[i])
		p.distanceFromOwned[i] = p.MSSP(was)
	}

	scores := make([]int64, len(p.AllEdges))
	for i, e := range p.AllEdges {
		scores[i] = p.getEdgeScore(e)
		if i != 0 {
			scores[i] += scores[i-1]
		}
	}

	if len(scores) == 0 || scores[len(scores)-1] == 0 {
		return p.MakePassMove()
	}

	score := rand.New(rand.NewSource(42)).Int63n(scores[len(scores)-1])
	for i, s := range scores {
		if score < s {
			e := &p.AllEdges[i]
			return p.MakeClaimMove(e.Src, e.Dst)
		}
	}

	panic("Inconsistent state")
	return p.MakePassMove()
}

func (p *Random0Player) Name() string { return "random0" }

func (p *Random0Player) getEdgeScore(e Edge) (score int64) {
	if e.Owner >= 0 {
		return
	}

	for mine := range p.Mines {
		rS := p.reachableFromMine[mine][e.Src]
		rD := p.reachableFromMine[mine][e.Dst]
		if rS && rD {
			continue
		}

		if !rS && !rD {
			ds := int64(p.Distance[mine][e.Src])
			so := int64(p.distanceFromOwned[mine][e.Src])
			if ds >= 0 {
				score += ds * ds / so
			}

			dd := int64(p.Distance[mine][e.Dst])
			do := int64(p.distanceFromOwned[mine][e.Dst])
			if dd >= 0 {
				score += dd * dd / do
			}
			continue
		}

		if rS {
			d := int64(p.Distance[mine][e.Dst])
			score += d * d
			score += p.countBonus(e.Dst, e.Src, mine)
		} else {
			d := int64(p.Distance[mine][e.Src])
			score += d * d
			score += p.countBonus(e.Src, e.Dst, mine)
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
