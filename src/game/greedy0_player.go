package game

type Greedy0Player struct {
	BaselinePlayer
}

func (p *Greedy0Player) MakeMove(moves []Move) Move {
	p.BaselinePlayer.PrepareForMove(moves)

	// Returns vertices (NOT sites), i.e. ints from the range [0..NumSites).
	// true on success, false on timeout (should not happen).
	u, v, ok := FindEdgeGreedy0(p)
	if !ok {
		return p.MakePassMove()
	}

	return p.MakeClaimMove(u, v)
}

func FindEdgeGreedy0(p *Greedy0Player) (int, int, bool) {
	reachable := make([]bool, p.NumSites)

	for _, v := range p.Mines {
		if !reachable[v] {
			dfsGreedy0(p, v, reachable)
		}
	}

	bestU, bestV, bestScore := -1, -1, 0
	for _, e := range p.AllEdges {
		if e.Owner >= 0 {
			continue
		}
		if !reachable[e.Src] && !reachable[e.Dst] {
			continue
		}
		if reachable[e.Src] && reachable[e.Dst] {
			if bestU < 0 {
				bestU, bestV = e.Src, e.Dst
			}
			continue
		}

		cur := 0

		upd := func(v int) {
			if !reachable[v] {
				for i := range p.Mines {
					d := p.Distance[i][v]
					cur += d * d
				}
			}
		}

		upd(e.Src)
		upd(e.Dst)

		if bestScore < cur {
			bestScore = cur
			bestU, bestV = e.Src, e.Dst
		}
	}

	if bestU >= 0 {
		return bestU, bestV, true
	}

	return 0, 0, false
}

func dfsGreedy0(p *Greedy0Player, v int, was []bool) {
	was[v] = true
	for _, eId := range p.Edges[v] {
		if p.AllEdges[eId].Owner != p.Punter {
			continue
		}
		u := p.AllEdges[eId].Dst
		if !was[u] {
			dfsGreedy0(p, u, was)
		}
	}
}

func (p *Greedy0Player) Name() string { return "greedy0" }
