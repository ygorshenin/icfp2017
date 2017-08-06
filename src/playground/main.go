package main

import (
	"bufio"
	"common"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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

func (g *graph) calcMineScore(u, player int, visited map[int]bool, sssp map[int]int) (score int64) {
	visited[u] = true
	score = int64(sssp[u]) * int64(sssp[u])

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

func (g *graph) calcFullScore(player int) (score int64) {
	for i, mine := range g.mines {
		visited := make(map[int]bool)
		score += g.calcMineScore(mine, player, visited, g.sssp[i])
	}
	return
}

func makeGraph(m *common.Map) (g graph) {
	numVertices := len(m.Sites)
	numEdges := len(m.Rivers)

	g.vertices = make([]int, numVertices)
	for i, site := range m.Sites {
		g.vertices[i] = site.Id
	}
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
var flagBots = flag.String("bots", "baseline,baseline", "Comma-separated list of bots")
var flagVisFile = flag.String("visfile", "", "filename to write visualizer information to")
var visWriter *bufio.Writer

func loadMap(path string) (m common.Map) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Can't read file:", path)
	}

	err = json.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatal("Can't parse map:", string(data))
	}
	return
}

func parseBots(s string) (bots []string) {
	parts := strings.Split(s, ",")
	for _, part := range parts {
		reps := strings.Split(part, "*")

		if len(reps) > 2 {
			log.Fatal("Invalid bots spec: " + s)
		}

		if len(reps) == 1 {
			bots = append(bots, reps[0])
			continue
		}

		n, err := strconv.Atoi(reps[1])
		if err != nil {
			log.Fatal("Invalid bots spec: " + s)
		}

		for i := 0; i < n; i++ {
			bots = append(bots, reps[0])
		}
	}
	return
}

func main() {
	flag.Parse()

	bots := parseBots(*flagBots)
	numPunters := len(bots)

	m := loadMap(*flagMap)

	if *flagVisFile != "" {
		visFile, err := os.Create(*flagVisFile)
		if err != nil {
			log.Fatal("Can't open vis file:", err)
		}
		visWriter = bufio.NewWriter(visFile)

		jsonMap, err := json.Marshal(&m)
		if err != nil {
			log.Fatal("Can't show map:", err)
		}
		fmt.Fprintln(visWriter, string(jsonMap))
	}

	punters := make([]common.PlayerProxy, numPunters)
	for i := range punters {
		punters[i] = common.MakePlayerProxy(bots[i])
		punters[i].Setup(i, numPunters, &m)
	}

	g := makeGraph(&m)

	moves := make([]common.Move, numPunters)
	for i := 0; i < numPunters; i++ {
		moves[i].Pass = &common.PassMove{Punter: i}
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

			if *flagVisFile != "" && move.Claim != nil {
				claim := move.Claim
				fmt.Fprintln(visWriter, claim.Punter, claim.Source, claim.Target)
			}

			if move.Pass != nil {
				numPasses[punter]++
			} else if move.Claim != nil {
				numPasses[punter] = 0
				curRivers++
				g.claimEdge(punter, move.Claim.Source, move.Claim.Target)
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
		log.Printf("Punter %v %v, score: %v", punter, punters[punter].Name(), score)
	}

	if *flagVisFile != "" {
		visWriter.Flush()
	}
}
