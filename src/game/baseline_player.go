package game

type BaselinePlayer struct {
	Graph

	Punter  int `json:"punter"`
	Punters int `json:"punters"`

	Settings Settings `json:"settings"`
	Futures  []Future `json:"futures"` // futures in the compressed format
	Passes   int      `json:"passes"`  // number of consecutive passes

	// Non-json fields are recalculated on every move.
	reachableFromMine [][]bool // reachableFromMine[i] is the reachability array from Mine i
	score             int64    // current score
	scores            []int64  // current scores for all punters
}

func (p *BaselinePlayer) MakeClaimMove(source, target int) Move {
	p.Passes = 0
	return MakeClaimMove(p.Punter, source, target)
}

func (p *BaselinePlayer) MakePassMove() Move {
	p.Passes++
	return MakePassMove(p.Punter)
}

func (p *BaselinePlayer) MakeSplurgeMove(route []int) Move {
	if !p.Settings.SplurgesMode {
		panic("cannot splurge: splurge mode is off")
	}
	if len(route) > p.Passes+2 {
		panic("not enough passes to splurge")
	}
	return MakeSplurgeMove(p.Punter, route)
}

func (p *BaselinePlayer) Setup(punter, punters int, m Map, s Settings) {
	p.Punter = punter
	p.Punters = punters
	p.Settings = s

	p.InitGraph(m)
}

func (p *BaselinePlayer) PrepareForMove(moves []Move) {
	p.ApplyMoves(moves)
	p.CalcReachabilityFromMines()
	p.CalcScores()
}

func (p *BaselinePlayer) MakeMove(moves []Move) Move {
	p.PrepareForMove(moves)

	// Returns vertices (NOT sites), i.e. ints from the range [0..NumSites).
	// true on success, false on timeout (should not happen).
	u, v, ok := p.FindEdge()
	if !ok {
		return p.MakePassMove()
	}
	return p.MakeClaimMove(u, v)
}

func (p *BaselinePlayer) Name() string {
	return "baseline"
}

func (p *BaselinePlayer) GetPunter() int {
	return p.Punter
}

func (p *BaselinePlayer) GetFutures() []Future {
	return p.Futures
}

func (p *BaselinePlayer) SetEdgeOwnership(a, b, owner int) {
	for _, eId := range p.Edges[a] {
		e := &p.AllEdges[eId]
		if e.Dst == b {
			if e.Owner >= 0 && e.Owner != owner {
				panic("a previously claimed edge was claimed in a non-pass move")
			}
			e.Owner = owner
			p.AllEdges[e.Id^1].Owner = owner
		}
	}
}

func (p *BaselinePlayer) ApplyMoves(moves []Move) {
	for _, m := range moves {
		if m.Type == Pass {
			continue
		}

		if m.Type == Claim {
			p.SetEdgeOwnership(m.Source, m.Target, m.Punter)
		}

		if m.Type == Splurge {
			for i := 0; i+1 < len(m.Route); i++ {
				p.SetEdgeOwnership(m.Route[i], m.Route[i+1], m.Punter)
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

func (p *BaselinePlayer) CalcScores() {
	p.scores = make([]int64, p.Punters)
	for pId := 0; pId < p.Punters; pId++ {
		for i := range p.Mines {
			var was []bool
			if pId == p.Punter {
				was = p.reachableFromMine[i]
			} else {
				was = make([]bool, p.NumSites)
				p.Dfs(p.Mines[i], pId, was)
			}
			for j := 0; j < p.NumSites; j++ {
				if was[j] {
					d := int64(p.Distance[i][j])
					p.scores[pId] += d * d
				}
			}
		}
	}
	p.score = p.scores[p.Punter]
}

// Returns the edge that results in the best increase in score.
func (p *BaselinePlayer) FindEdge() (int, int, bool) {
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
