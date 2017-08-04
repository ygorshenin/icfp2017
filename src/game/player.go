package game

type Player struct {
	Punter  int
	Punters int
	Map     Map
}

func (p *Player) makeClaimMove(source, target int) Move {
	return MakeClaimMove(p.Punter, source, target)
}

func (p *Player) makePassMove() Move {
	return MakePassMove(p.Punter)
}

func (p *Player) Setup(punter, punters int, m Map) {
	p.Punter = punter
	p.Punters = punters
	p.Map = m
}

func (p *Player) MakeMove(moves []Move) Move {
	return p.makePassMove()
}
