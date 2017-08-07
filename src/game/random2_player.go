package game

type Random2Player struct {
	Random1Player
}

func (p *Random2Player) Setup(punter, punters int, m Map, s Settings) {
	p.Random1Player.Setup(punter, punters, m, s)
	p.setupFutures()
}

func (p *Random2Player) Name() string {
	return "random2"
}

func (p *Random2Player) setupFutures() {
	if !p.Settings.FuturesMode {
		return
	}

	maxU := make([]int, len(p.Mines))
	maxD := make([]int, len(p.Mines))

	for i := range p.Mines {
		maxU[i] = -1
		maxD[i] = -1
	}

	for u := 0; u < p.NumSites; u++ {
		bestI := -1
		bestD := p.NumSites
		for i, m := range p.Mines {
			if u == m {
				continue
			}
			d := p.Distance[i][u]
			if d >= 0 && d < bestD {
				bestI = i
				bestD = d
			}
		}

		if bestI >= 0 {
			if maxD[bestI] < bestD {
				maxU[bestI] = u
				maxD[bestI] = bestD
			}
		}
	}

	for i, m := range p.Mines {
		if maxU[i] >= 0 {
			p.Futures = append(p.Futures, Future{Src: m, Dst: maxU[i]})
		}
	}
}
