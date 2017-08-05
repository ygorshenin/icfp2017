package game

type BaselinePlayer struct {
	Graph

	Punter  int `json:"punter"`
	Punters int `json:"punters"`
	Map     Map `json:"map"`

	SiteToIndex map[int]int `json:"siteToIndex"` // maps sites to the range [0..NumSites)
	IndexToSite []int       `json:"indexToSite"` // ...and back

	// Non-json fields are recalculated on every move.
	reachableFromMine [][]bool // reachableFromMine[i] is the reachability array from Mine i
	score             int64    // current score
}

func (p *BaselinePlayer) makeClaimMove(source, target int) Move {
	return MakeClaimMove(p.Punter, source, target)
}

func (p *BaselinePlayer) makePassMove() Move {
	return MakePassMove(p.Punter)
}

func (p *BaselinePlayer) Setup(punter, punters int, m Map) {
	p.Punter = punter
	p.Punters = punters
	p.Map = m

	p.NumSites = 0
	p.SiteToIndex = make(map[int]int)
	for _, s := range m.Sites {
		_, ok := p.SiteToIndex[s.Id]
		if !ok {
			p.SiteToIndex[s.Id] = p.NumSites
			p.NumSites++
		}
	}
	p.IndexToSite = make([]int, p.NumSites)
	for k, v := range p.SiteToIndex {
		p.IndexToSite[v] = k
	}

	p.Mines = make([]int, len(m.Mines))
	for i, id := range m.Mines {
		p.Mines[i] = p.SiteToIndex[id]
	}

	p.AllEdges = make([]Edge, 2*len(m.Rivers))
	p.Edges = make([][]int, p.NumSites)
	for i, r := range m.Rivers {
		a := p.SiteToIndex[r.Source]
		b := p.SiteToIndex[r.Target]

		// todo(@m) use degs
		p.AllEdges[2*i] = Edge{Id: 2 * i, Src: a, Dst: b, Owner: -1}
		p.AllEdges[2*i+1] = Edge{Id: 2*i + 1, Src: b, Dst: a, Owner: -1}
		p.Edges[a] = append(p.Edges[a], 2*i)
		p.Edges[b] = append(p.Edges[b], 2*i+1)
	}

	p.InitShortestPaths()
}

func (p *BaselinePlayer) PrepareForMove(moves []Move) {
	p.ApplyMoves(moves)
	p.CalcReachabilityFromMines()
	p.CalcScore()
}

func (p *BaselinePlayer) MakeMove(moves []Move) Move {
	p.PrepareForMove(moves)

	// Returns vertices (NOT sites), i.e. ints from the range [0..NumSites).
	// true on success, false on timeout (should not happen).
	u, v, ok := FindEdgeVerySimple(p)
	if !ok {
		return p.makePassMove()
	}
	u = p.IndexToSite[u]
	v = p.IndexToSite[v]
	return p.makeClaimMove(u, v)
}

func (p *BaselinePlayer) Name() string {
	return "baseline"
}

func (p *BaselinePlayer) ApplyMoves(moves []Move) {
	for _, m := range moves {
		if m.Type == Pass {
			continue
		}

		a := p.SiteToIndex[m.Source]
		b := p.SiteToIndex[m.Target]
		o := m.Punter

		for _, eId := range p.Edges[a] {
			e := &p.AllEdges[eId]
			if e.Dst == b {
				if e.Owner >= 0 && e.Owner != o {
					panic("a previously claimed edge was claimed in a non-pass move")
				}
				e.Owner = o
				p.AllEdges[e.Id^1].Owner = o
			}
		}
	}
}

func (p *BaselinePlayer) CalcReachabilityFromMines() {
	p.reachableFromMine = make([][]bool, len(p.Mines))
	for i, s := range p.Mines {
		p.reachableFromMine[i] = make([]bool, p.NumSites)
		p.Graph.Dfs(s, p.Punter, p.reachableFromMine[i])
	}
}

func (p *BaselinePlayer) CalcScore() {
	p.score = 0
	for i := range p.Mines {
		for j := 0; j < p.NumSites; j++ {
			if p.reachableFromMine[i][j] {
				d := int64(p.Distance[i][j])
				p.score += d * d
			}
		}
	}
}

// Returns the edge that results in the best increase in score.
func FindEdgeVerySimple(p *BaselinePlayer) (int, int, bool) {
	bestU, bestV, bestInc := -1, -1, int64(0)

	for _, e := range p.AllEdges {
		if e.Owner >= 0 {
			continue
		}

		var curInc int64
		for i := range p.Mines {
			rS := p.reachableFromMine[i][e.Src]
			rD := p.reachableFromMine[i][e.Dst]
			if rS == rD {
				continue
			}
			d := int64(p.Distance[i][e.Src])
			if rS {
				d = int64(p.Distance[i][e.Dst])
			}
			curInc += d * d
		}

		if bestInc < curInc {
			bestInc = curInc
			bestU, bestV = e.Src, e.Dst
		}
	}

	if bestU >= 0 {
		return bestU, bestV, true
	}

	return 0, 0, false
}
