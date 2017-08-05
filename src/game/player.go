package game

type Player interface {
	Setup(punter, punters int, m Map)
	MakeMove(moves []Move) Move
	Name() string
}

func MakePlayer(name string) Player {
	if name == "zombie" {
		return new(ZombiePlayer)
	}

	if name == "baseline" {
		return new(BaselinePlayer)
	}

	panic("Unknown name: " + name)
}
