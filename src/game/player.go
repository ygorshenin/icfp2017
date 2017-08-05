package game

type Player struct {
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

func (p *Player) makeClaimMove(source, target int) Move {
	return MakeClaimMove(p.Punter, source, target)
}

func (p *Player) makePassMove() Move {
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

func (st *Player) Setup(punter, punters int, m Map) {
	st.Punter = punter
	st.Punters = punters
	st.Map = m

	st.NumSites = 0
	st.SiteToIndex = make(map[int]int)
	for _, s := range m.Sites {
		_, ok := st.SiteToIndex[s]
		if !ok {
			st.SiteToIndex[s] = st.NumSites
			st.NumSites++
		}
	}
	st.IndexToSite = make([]int, st.NumSites)
	for k, v := range st.SiteToIndex {
		st.IndexToSite[v] = k
	}

	st.Mines = make([]int, len(m.Mines))
	for i, id := range m.Mines {
		st.Mines[i] = st.SiteToIndex[id]
	}

	st.AllEdges = make([]Edge, 2*len(m.Rivers))
	st.Edges = make([][]int, st.NumSites)
	for i, r := range m.Rivers {
		a := st.SiteToIndex[r.Source]
		b := st.SiteToIndex[r.Target]

		// todo(@m) use degs
		st.AllEdges[2*i] = Edge{Id: 2 * i, Src: a, Dst: b, Owner: -1}
		st.AllEdges[2*i+1] = Edge{Id: 2*i + 1, Src: b, Dst: a, Owner: -1}
		st.Edges[a] = append(st.Edges[a], 2*i)
		st.Edges[b] = append(st.Edges[b], 2*i+1)
	}

	st.Distance = make([][]int, len(m.Mines))
	for i := range st.Distance {
		st.Distance[i] = calcShortestPath(st.Mines[i], st.AllEdges, st.Edges)
	}
}

func (p *Player) MakeMove(moves []Move) Move {
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

func (st *Player) ApplyMoves(moves []Move) {
	for _, m := range moves {
		if m.Type == Pass {
			continue
		}

		a := st.SiteToIndex[m.Source]
		b := st.SiteToIndex[m.Target]
		o := m.Punter

		for _, eId := range st.Edges[a] {
			e := &st.AllEdges[eId]
			if e.Dst == b {
				if e.Owner >= 0 && e.Owner != o {
					panic("a previously claimed edge was claimed in a non-pass move")
				}
				e.Owner = o
				st.AllEdges[e.Id^1].Owner = o
			}
		}
	}
}

func FindEdgeVerySimple(st *Player) (int, int, bool) {
	reachable := make([]bool, st.NumSites)

	for _, v := range st.Mines {
		if !reachable[v] {
			dfsVerySimple(st, v, reachable)
		}
	}

	bestU, bestV, bestScore := -1, -1, 0
	for _, e := range st.AllEdges {
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
				for i := range st.Mines {
					d := st.Distance[i][v]
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

func dfsVerySimple(st *Player, v int, was []bool) {
	was[v] = true
	for _, eId := range st.Edges[v] {
		if st.AllEdges[eId].Owner != st.Punter {
			continue
		}
		u := st.AllEdges[eId].Dst
		if !was[u] {
			dfsVerySimple(st, u, was)
		}
	}
}
