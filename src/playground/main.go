package main

import (
	"common"
	"encoding/json"
	"flag"
	"game"
	"io/ioutil"
	"log"
)

type edge struct {
	from  int
	to    int
	owner int
}

type graph struct {
	vertices []int
	mines    []int
	sssp     map[int]map[int]int
	edges    []edge
	adj      map[int][]int
}

func (g *graph) addEdge(i int, e edge) {
	g.edges[i] = e

	_, ok := g.adj[e.from]
	if !ok {
		g.adj[e.from] = make([]int, 0)
	}

	g.adj[e.from] = append(g.adj[e.from], i)
}

func (g *graph) claimEdge(owner, from, to int) {
	out := g.adj[from]
	for _, e := range out {
		edge := &g.edges[e]
		revEdge := &g.edges[e^1]
		if edge.to == to {
			if edge.from != from || revEdge.from != to || revEdge.to != from {
				panic("Inconsistent state")
			}

			if edge.owner >= 0 || revEdge.owner >= 0 {
				panic("Edge alreay claimed!")
			}

			edge.owner = owner
			revEdge.owner = owner
		}
	}
}

func (g *graph) bfs(root int, sssp map[int]int) {
	n := len(g.vertices)

	queue := make([]int, n)
	head, tail := 0, 0

	sssp[root] = 0
	queue[tail] = root
	tail++

	for head < tail {
		u := queue[head]
		head++

		out, ok := g.adj[u]
		if !ok {
			continue
		}

		for _, e := range out {
			edge := &g.edges[e]

			if _, ok := sssp[edge.to]; !ok {
				sssp[edge.to] = sssp[edge.from] + 1
				queue[tail] = edge.to
				tail++
			}
		}
	}
}

func (g *graph) calcMineScore(u, player int, visited map[int]bool, sssp map[int]int) (score int) {
	visited[u] = true
	score = sssp[u] * sssp[u]

	out, ok := g.adj[u]
	if !ok {
		return
	}

	for _, e := range out {
		edge := &g.edges[e]
		if edge.owner != player {
			continue
		}

		vis, ok := visited[edge.to]
		if ok && vis {
			continue
		}

		score += g.calcMineScore(edge.to, player, visited, sssp)
	}

	return
}

func (g *graph) calcFullScore(player int) (score int) {
	for i, mine := range g.mines {
		visited := make(map[int]bool)
		score += g.calcMineScore(mine, player, visited, g.sssp[i])
	}
	return
}

func makeGraph(m *game.Map) (g graph) {
	numEdges := len(m.Rivers)

	g.vertices = m.Sites
	g.mines = m.Mines
	g.edges = make([]edge, numEdges*2)
	g.adj = make(map[int][]int)

	for i, river := range m.Rivers {
		g.addEdge(2*i, edge{from: river.Source, to: river.Target, owner: -1})
		g.addEdge(2*i+1, edge{from: river.Target, to: river.Source, owner: -1})
	}

	g.sssp = make(map[int]map[int]int)
	for i, mine := range g.mines {
		g.sssp[i] = make(map[int]int)
		g.bfs(mine, g.sssp[i])
	}
	return g
}

const MaxPasses = 10

var flagMap = flag.String("map", "", "Path to a JSON-encoded map")
var flagPunters = flag.Int("punters", 2, "Number of bots")

func loadMap(path string) game.Map {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Can't read file:", path)
	}

	var m common.Map
	err = json.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatal("Can't parse map:", data)
	}
	return common.ToGameMap(&m)
}

func main() {
	flag.Parse()

	numPunters := *flagPunters

	m := loadMap(*flagMap)
	punters := make([]game.Player, numPunters)
	for i := range punters {
		punters[i].Setup(i, *flagPunters, m)
	}

	g := makeGraph(&m)

	moves := make([]game.Move, numPunters)
	for i := 0; i < numPunters; i++ {
		moves[i] = game.MakePassMove(i)
	}

	zombies := make([]bool, numPunters)
	numPasses := make([]int, numPunters)

	numRivers := len(m.Rivers)
	curRivers, numZombies := 0, 0
	for turn := 0; curRivers != numRivers && numZombies != numPunters; turn++ {
		for punter := 0; punter < numPunters && curRivers != numRivers && numZombies != numPunters; punter++ {
			if zombies[punter] {
				continue
			}

			move := punters[punter].MakeMove(moves)

			log.Println("Move: ", move.String())

			switch move.Type {
			case game.Pass:
				numPasses[punter]++
			case game.Claim:
				numPasses[punter] = 0
				curRivers++
				g.claimEdge(punter, move.Source, move.Target)
			}

			if numPasses[punter] == MaxPasses {
				zombies[punter] = true
				numZombies++
			}

			moves[punter] = move
		}
	}

	for punter := 0; punter < numPunters; punter++ {
		score := g.calcFullScore(punter)
		log.Printf("Punter %v, score: %v", punter, score)
	}
}
