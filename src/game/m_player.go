package game

type MPlayer struct {
	BaselinePlayer
}

func (p *MPlayer) Setup(punter, punters int, m Map, s Settings) {
	p.BaselinePlayer.Setup(punter, punters, m, s)

	p.SetupFutures()
}

func (p *MPlayer) SetupFutures() {
	if !p.Settings.FuturesMode {
		return
	}
	maxDist := p.NumSites / 5
	if maxDist > 15 {
		maxDist = 15
	}
	for i, m := range p.Mines {
		best, bestDist := -1, -1
		for j := 0; j < p.NumSites; j++ {
			if j == m {
				continue
			}
			d := p.Distance[i][j]
			//			if d > maxDist {
			//				continue
			//			}
			if best < 0 || bestDist < d {
				best, bestDist = j, d
			}
		}
		if best >= 0 {
			p.Futures = append(p.Futures, Future{Src: m, Dst: best})
		}
	}
}

func (p *MPlayer) MakeMove(moves []Move) Move {
	p.BaselinePlayer.PrepareForMove(moves)

	u, v, ok := p.BaselinePlayer.FindEdge()
	if !ok {
		return p.MakePassMove()
	}
	return p.MakeClaimMove(u, v)
}

func (p *MPlayer) Name() string {
	return "m"
}

func (p *MPlayer) GetFutures() []Future {
	return p.Futures
}
