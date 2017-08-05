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
		g.Distance[i] = g.CalcShortestPaths(g.Mines[i])
	}
}

func (g *Graph) CalcShortestPaths(s int) []int {
	n := len(g.Edges)
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
		for _, eId := range g.Edges[v] {
			u := g.AllEdges[eId].Dst
			if d[u] > 1+d[v] {
				d[u] = 1 + d[v]
				q[qh] = u
				qh++
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
