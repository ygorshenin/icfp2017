package game

type ZombiePlayer struct {
	Punter int `json:"punter"`
}

func (p *ZombiePlayer) Setup(punter, punters int, m Map, s Settings) {
	p.Punter = punter
}

func (p *ZombiePlayer) MakeMove(moves []Move) Move {
	return MakePassMove(p.Punter)
}

func (p *ZombiePlayer) Name() string { return "zombie" }

func (p *ZombiePlayer) GetPunter() int { return p.Punter }

func (p *ZombiePlayer) GetFutures() []Future { return nil }
