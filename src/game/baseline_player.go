package game

type BaselinePlayer struct {
	Punter  int `json:"punter"`
	Punters int `json:"punters"`
	Map     Map `json:"map"`

	Distance [][]int `json:"distance"` // distance[i][j] = shortest distance from mine i to site j

	AllEdges []Edge  `json:"allEdges"`
	Edges    [][]int `json:"edges"`

	NumSites    int         `json:"numSites"`
	SiteToIndex map[int]int `json:"siteToIndex"` // maps sites to the range [0..NumSites)
	IndexToSite []int       `json:"indexToSite"` // ...and back

	Mines []int `json:"mines"` // indexes of mines
}

type Edge struct {
	Id    int `json:"id"`
	Src   int `json:"src"`
	Dst   int `json:"dst"`
	Owner int `json:"owner"`
}

func (p *BaselinePlayer) makeClaimMove(source, target int) Move {
	return MakeClaimMove(p.Punter, source, target)
}

func (p *BaselinePlayer) makePassMove() Move {
	return MakePassMove(p.Punter)
}

func calcShortestPath(s int, allEdges []Edge, edges [][]int) []int {
	n := len(edges)
	d := make([]int, n)
	for i := range d {
		d[i] = n + 1
	}
	d[s] = 0
	q := make([]int, n)
	qt, qh := 0, 1
	q[0] = s
	for qt < qh {
		v := q[qt]
		qt++
		for _, eId := range edges[v] {
			u := allEdges[eId].Dst
			if d[u] > 1+d[v] {
				d[u] = 1 + d[v]
				q[qh] = u
				qh++
			}
		}
	}
	return d
}

func (p *BaselinePlayer) Setup(punter, punters int, m Map) {
	p.Punter = punter
	p.Punters = punters
	p.Map = m

	p.NumSites = 0
	p.SiteToIndex = make(map[int]int)
	for _, s := range m.Sites {
		_, ok := p.SiteToIndex[s]
		if !ok {
			p.SiteToIndex[s] = p.NumSites
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

	p.Distance = make([][]int, len(m.Mines))
	for i := range p.Distance {
		p.Distance[i] = calcShortestPath(p.Mines[i], p.AllEdges, p.Edges)
	}
}

func (p *BaselinePlayer) MakeMove(moves []Move) Move {
	p.ApplyMoves(moves)

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

func FindEdgeVerySimple(p *BaselinePlayer) (int, int, bool) {
	reachable := make([]bool, p.NumSites)

	for _, v := range p.Mines {
		if !reachable[v] {
			dfsVerySimple(p, v, reachable)
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

func dfsVerySimple(p *BaselinePlayer, v int, was []bool) {
	was[v] = true
	for _, eId := range p.Edges[v] {
		if p.AllEdges[eId].Owner != p.Punter {
			continue
		}
		u := p.AllEdges[eId].Dst
		if !was[u] {
			dfsVerySimple(p, u, was)
		}
	}
}
