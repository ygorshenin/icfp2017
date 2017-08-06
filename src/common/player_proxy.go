package common

import (
	"game"
	"log"
)

type PlayerProxy struct {
	Player game.Player     `json:"player"`
	Index  CompressedIndex `json:"index"`
}

func (pp *PlayerProxy) toGameMove(move *Move) game.Move {
	if move.Pass != nil {
		return game.MakePassMove(move.Pass.Punter)
	}
	if move.Splurge != nil {
		route := make([]int, len(move.Splurge.Route))
		for i, v := range len(move.Splurge.Route) {
			route[i] = pp.Index.Forward[v]
		}
		return game.MakeSplurgeMove(move.Splurge.Punter, route)
	}
	claim := move.Claim
	return game.MakeClaimMove(claim.Punter, pp.Index.Forward[claim.Source], pp.Index.Forward[claim.Target])
}

func (pp *PlayerProxy) toGameMoves(moves []Move) (gmoves []game.Move) {
	n := len(moves)
	gmoves = make([]game.Move, n, n)
	for i, m := range moves {
		gmoves[i] = pp.toGameMove(&m)
	}
	return
}

func (pp *PlayerProxy) fromGameMove(m *game.Move) (r Move) {
	switch m.Type {
	case game.Claim:
		r.Claim = &ClaimMove{
			Punter: m.Punter,
			Source: pp.Index.Backward[m.Source],
			Target: pp.Index.Backward[m.Target]}
	case game.Pass:
		r.Pass = &PassMove{Punter: m.Punter}
	case game.Splurge:
		route := make([]int, len(m.Route))
		for i, v := range m.Route {
			route[i] = pp.Index.Backward[v]
		}
		r.Splurge = &SplurgeMove{
			Punter: m.Punter,
			Route:  route}
	default:
		log.Fatal("Unknown move type:", m.Type)
	}
	r.State = pp
	return
}

func (pp *PlayerProxy) Setup(punter, punters int, m *Map, settings game.Settings) {
	allSites := make([]int, len(m.Sites))
	for i, site := range m.Sites {
		allSites[i] = site.Id
	}
	pp.Index.Setup(allSites)

	var gm game.Map

	gm.Sites = make([]int, len(pp.Index.Backward))
	for i, site := range m.Sites {
		gm.Sites[i] = pp.Index.Forward[site.Id]
	}

	gm.Rivers = make([]game.River, len(m.Rivers))
	for i, river := range m.Rivers {
		gm.Rivers[i].Source = pp.Index.Forward[river.Source]
		gm.Rivers[i].Target = pp.Index.Forward[river.Target]
	}

	gm.Mines = make([]int, len(m.Mines))
	for i, mine := range m.Mines {
		gm.Mines[i] = pp.Index.Forward[mine]
	}

	pp.Player.Setup(punter, punters, gm, settings)
}

func (pp *PlayerProxy) MakeMove(moves []Move) Move {
	gm := pp.Player.MakeMove(pp.toGameMoves(moves))
	return pp.fromGameMove(&gm)
}

func (pp *PlayerProxy) Name() string {
	return pp.Player.Name()
}

func (pp *PlayerProxy) GetPunter() int {
	return pp.Player.GetPunter()
}

func (pp *PlayerProxy) GetFutures() []game.Future {
	fs := pp.Player.GetFutures()
	res := make([]game.Future, len(fs))
	for i, f := range fs {
		res[i].Src = pp.Index.Backward[f.Src]
		res[i].Dst = pp.Index.Backward[f.Dst]
	}
	return res
}

func MakePlayerProxy(name string) (pp PlayerProxy) {
	pp.Player = game.MakePlayer(name)
	return
}
