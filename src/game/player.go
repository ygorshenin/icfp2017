package game

type Player interface {
	Setup(punter, punters int, m Map)
	MakeMove(moves []Move) Move
	Name() string
}

func MakePlayer(name string) Player {
	switch name {
	case "zombie":
		return new(ZombiePlayer)
	case "baseline":
		return new(BaselinePlayer)
	case "greedy0":
		return new(Greedy0Player)
	}
	panic("Unknown name: " + name)
}
