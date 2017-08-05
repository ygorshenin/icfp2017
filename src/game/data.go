package game

import "fmt"

type River struct {
	Source int `json:"source"`
	Target int `json:"target"`
}

type Map struct {
	Sites  []int
	Rivers []River
	Mines  []int
}

type MoveType int

const (
	Claim = iota
	Pass
)

type Move struct {
	Type   MoveType
	Punter int
	Source int
	Target int
}

func (mt MoveType) String() string {
	switch mt {
	case Claim:
		return "Claim"
	case Pass:
		return "Pass"
	}
	return "Unknown move"
}

func (m Move) String() string {
	switch m.Type {
	case Claim:
		return fmt.Sprintf("Punter=%v, River=(%v,%v)", m.Punter, m.Source, m.Target)
	case Pass:
		return fmt.Sprintf("%t", m.Type)
	}
	return "Bad Move"
}

func MakeClaimMove(punter, source, target int) Move {
	return Move{Type: Claim, Punter: punter, Source: source, Target: target}
}

func MakePassMove(punter int) Move {
	return Move{Type: Pass, Punter: punter, Source: 0, Target: 0}
}
