package main

import (
	"bufio"
	"common"
	"encoding/json"
	"flag"
	"fmt"
	"game"
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
	isMine   map[int]bool
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
	found := false
	for _, e := range out {
		edge := &g.edges[e]
		revEdge := &g.edges[e^1]
		if edge.to == to {
			if found || edge.from != from || revEdge.from != to || revEdge.to != from {
				panic("Inconsistent state")
			}

			if edge.owner >= 0 || revEdge.owner >= 0 {
				panic("Edge alreay claimed!")
			}

			edge.owner = owner
			revEdge.owner = owner
			found = true
		}
	}
	if !found {
		panic("Can't claim unknown edge: " + strconv.Itoa(from) + " " + strconv.Itoa(to))
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

func (g *graph) dfs(u int, player int, visited map[int]bool) {
	visited[u] = true

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

		g.dfs(edge.to, player, visited)
	}
}

func (g *graph) calcFullScore(player int, futures [][2]int, s game.Settings) (score int64) {
	for _, mine := range g.mines {
		visited := make(map[int]bool)
		score += g.calcMineScore(mine, player, visited, g.sssp[mine])
	}

	if !s.FuturesMode && len(futures) > 0 {
		log.Println("Warning: futures mode is OFF, but the player thinks it's ON")
	}

	if s.FuturesMode {
		for _, f := range futures {
			a, b := f[0], f[1]
			if im, ok := g.isMine[a]; !ok || !im {
				log.Fatal("A future's starting point is not a mine", a, b)
			}
			visited := make(map[int]bool)
			g.dfs(a, player, visited)
			d := int64(g.sssp[a][b])
			d3 := d * d * d

			if visited[b] {
				log.Println("Punter ", player, " satisfied future, bonus: ", d3)
				score += d3
			} else {
				log.Println("Punter ", player, " failed future, penalty: ", d3)
				score -= d3
			}
		}
	}

	return
}

// Computes upper bound on the score for any player, without futures
func (g *graph) scoreUpperBound() (score int64) {
	for _, mine := range g.mines {
		for _, vertex := range g.vertices {
			d := int64(g.sssp[mine][vertex])
			score += d * d
		}
	}
	return
}

func (g *graph) futureUpperBound() int64 {
	var score int64
	for _, mine := range g.mines {
		for _, vertex := range g.vertices {
			d := int64(g.sssp[mine][vertex])
			if d > score {
				score = d
			}
		}
	}
	return score * score * score
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

	g.isMine = make(map[int]bool)
	for _, m := range m.Mines {
		g.isMine[m] = true
	}

	for i, river := range m.Rivers {
		g.addEdge(2*i, edge{from: river.Source, to: river.Target, owner: -1})
		g.addEdge(2*i+1, edge{from: river.Target, to: river.Source, owner: -1})
	}

	g.sssp = make(map[int]map[int]int)
	for _, mine := range g.mines {
		g.sssp[mine] = make(map[int]int)
		g.bfs(mine, g.sssp[mine])
	}
	return g
}

const MaxPasses = 10

var flagMap = flag.String("map", "", "Path to a JSON-encoded map")
var flagBots = flag.String("bots", "baseline,baseline", "Comma-separated list of bots")
var flagVisFile = flag.String("visfile", "", "filename to write visualizer information to")
var flagSettings = flag.String("settings", "", "Comma-separated list of settings")
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

func parseSettings(str string) (s game.Settings) {
	if str == "" {
		return
	}
	parts := strings.Split(str, ",")
	for _, part := range parts {
		switch part {
		case "futures":
			s.FuturesMode = true
		case "splurges":
			s.SplurgesMode = true
		default:
			log.Fatal("Bad value of settings:", str, ", can't read", part)
		}
	}
	return
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	bots := parseBots(*flagBots)
	numPunters := len(bots)

	m := loadMap(*flagMap)

	settings := parseSettings(*flagSettings)
	log.Println("Settings:", settings)

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
	futures := make([][][2]int, numPunters)
	for i := range punters {
		punters[i] = common.MakePlayerProxy(bots[i])
		punters[i].Setup(i, numPunters, &m, settings)

		fs := punters[i].GetFutures()
		futures[i] = make([][2]int, len(fs))
		for j, f := range fs {
			futures[i][j][0] = f.Src
			futures[i][j][1] = f.Dst
		}
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
			} else if move.Splurge != nil {
				if !settings.SplurgesMode || numPasses[punter]+1 < len(move.Splurge.Route) {
					// Cannot splurge, pass.
					numPasses[punter]++
				} else {
					for i := 0; i+1 < len(move.Splurge.Route); i++ {
						u := move.Splurge.Route[i]
						v := move.Splurge.Route[i+1]
						g.claimEdge(punter, u, v)
					}
					numPasses[punter] = 0
				}
			}

			if numPasses[punter] == MaxPasses {
				zombies[punter] = true
				numZombies++
			}

			moves[punter] = move
		}
	}

	scores := make([]int64, numPunters)
	var maxScore int64
	for punter := 0; punter < numPunters; punter++ {
		scores[punter] = g.calcFullScore(punter, futures[punter], settings)
		if scores[punter] > maxScore {
			maxScore = scores[punter]
		}
	}

	sub := g.scoreUpperBound()
	fub := g.futureUpperBound()
	log.Printf("Score upper bound (no futures): %v", sub)
	log.Printf("Future upper bound: %v", fub)

	for punter, score := range scores {
		fr := float64(score) * 100 / float64(sub)
		if score == maxScore {
			log.Printf("* Punter %v %v, score: %v (%.2f%%)", punter, punters[punter].Name(), score, fr)
		} else {
			log.Printf("  Punter %v %v, score: %v (%.2f%%)", punter, punters[punter].Name(), score, fr)
		}
	}

	if *flagVisFile != "" {
		visWriter.Flush()
	}
}
