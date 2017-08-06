package game

type Edge struct {
	Id    int `json:"id"`
	Src   int `json:"src"`
	Dst   int `json:"dst"`
	Owner int `json:"owner"`
}

type Graph struct {
	NumSites int     `json:"numSites"`
	AllEdges []Edge  `json:"allEdges"`
	Edges    [][]int `json:"edges"`
	Mines    []int   `json:"mines"`    // indexes of mines
	Distance [][]int `json:"distance"` // distance[i][j] = shortest distance from mine i to site j
}

func (g *Graph) InitShortestPaths() {
	g.Distance = make([][]int, len(g.Mines))
	for i := range g.Distance {
		g.Distance[i] = g.SSSP(g.Mines[i])
	}
}

func (g *Graph) SSSP(s int) []int {
	n := len(g.Edges)
	d := make([]int, n)
	for i := range d {
		d[i] = -1
	}
	d[s] = 0
	q := make([]int, n)
	qt, qh := 0, 1
	q[0] = s
	for qt < qh {
		v := q[qt]
		qt++
		for _, eId := range g.Edges[v] {
			u := g.AllEdges[eId].Dst
			if d[u] < 0 || d[u] > 1+d[v] {
				d[u] = 1 + d[v]
				q[qh] = u
				qh++
			}
		}
	}
	return d
}

func (g *Graph) MSSP(was []bool) []int {
	n := len(g.Edges)
	q := make([]int, n)
	qh, qt := 0, 0
	d := make([]int, n)
	for i := range d {
		if was[i] {
			d[i] = 0
			q[qt] = i
			qt++
		} else {
			d[i] = -1
		}
	}

	for qh < qt {
		u := q[qh]
		qh++
		for _, eId := range g.Edges[u] {
			v := g.AllEdges[eId].Dst
			if d[v] < 0 || d[v] > 1+d[u] {
				d[v] = 1 + d[u]
				q[qt] = v
				qt++
			}
		}
	}

	return d
}

func (g *Graph) Dfs(u, owner int, was []bool) {
	was[u] = true
	for _, eId := range g.Edges[u] {
		e := &g.AllEdges[eId]
		if e.Owner != owner {
			continue
		}
		v := e.Dst
		if !was[v] {
			g.Dfs(v, owner, was)
		}
	}
}
