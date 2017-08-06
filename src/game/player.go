package game

type Player interface {
	Setup(punter, punters int, m Map)
	MakeMove(moves []Move) Move
	Name() string
	GetPunter() int
}

func MakePlayer(name string) Player {
	switch name {
	case "zombie":
		return new(ZombiePlayer)
	case "baseline":
		return new(BaselinePlayer)
	case "greedy0":
		return new(Greedy0Player)
	case "random0":
		return new(Random0Player)
	}
	panic("Unknown name: " + name)
}
