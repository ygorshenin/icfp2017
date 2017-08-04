package game

type River struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

type Map struct {
	Sites  []int
	Rivers []River
	Mines  []int
}

const (
	Claim = iota
	Pass  = iota
)

type Move struct {
	Type   int
	Punter int
	Source int
	Target int
}

func MakeClaimMove(punter, source, target int) Move {
	return Move{Type: Claim, Punter: punter, Source: source, Target: target}
}

func MakePassMove(punter int) Move {
	return Move{Type: Pass, Punter: punter, Source: 0, Target: 0}
}
