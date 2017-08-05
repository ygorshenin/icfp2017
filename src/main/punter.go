package main

import (
	"bufio"
	"encoding/json"
	"game"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	name = "MIPT Lambda"
)

type Me struct {
	Me string `json:"me"`
}

type You struct {
	You string `json:"you"`
}

type Site struct {
	Id int `json:"id"`
}

// The Map structure from server is inconvinient, this is a wrapper.
type Map struct {
	Sites  []Site       `json:"sites"`
	Rivers []game.River `json:"rivers"`
	Mines  []int        `json:"mines"`
}

func toGameMap(m *Map) game.Map {
	sites := make([]int, len(m.Sites), len(m.Sites))
	for i, site := range m.Sites {
		sites[i] = site.Id
	}
	return game.Map{Sites: sites, Rivers: m.Rivers, Mines: m.Mines}
}

type Ready struct {
	Ready int          `json:"ready"`
	State *game.Player `json:"state"`
}

type ClaimMove struct {
	Punter int `json:"punter"`
	Source int `json:"source"`
	Target int `json:"target"`
}

type PassMove struct {
	Punter int `json:"punter"`
}

type Move struct {
	Claim *ClaimMove   `json:"claim,omitempty"`
	Pass  *PassMove    `json:"pass,omitempty"`
	State *game.Player `json:"state"`
}

type Moves struct {
	Moves []Move `json:"moves"`
}

func toGameMove(m *Move) game.Move {
	if m.Pass != nil {
		return game.MakePassMove(m.Pass.Punter)
	}
	claim := m.Claim
	return game.MakeClaimMove(claim.Punter, claim.Source, claim.Target)
}

func fromGameMove(m *game.Move, p *game.Player) (r Move) {
	switch m.Type {
	case game.Claim:
		r.Claim = &ClaimMove{Punter: m.Punter, Source: m.Source, Target: m.Target}
	case game.Pass:
		r.Pass = &PassMove{Punter: m.Punter}
	default:
		log.Fatal("Unknown move type:", m.Type)
	}
	r.State = p
	return
}

func toGameMoves(moves []Move) (r []game.Move) {
	n := len(moves)
	r = make([]game.Move, n, n)
	for i, m := range moves {
		r[i] = toGameMove(&m)
	}
	return
}

type Score struct {
	Punter int `json:"punter"`
	Score  int `json:"score"`
}

type Stop struct {
	Moves  []Move  `json:"move"`
	Scores []Score `json:"scores"`
}

type Step struct {
	Punter  *int `json:"punter"`
	Punters *int `json:"punters"`
	Map     *Map `json:"map"`

	Moves *Moves       `json:"move"`
	Stop  *Stop        `json:"stop"`
	State *game.Player `json:"state"`
}

func sendMessage(w *bufio.Writer, message interface{}) {
	bs, err := json.Marshal(message)
	if err != nil {
		log.Fatal("Can't send message:", err)
	}
	ss := string(bs)

	io.WriteString(w, strconv.Itoa(len(ss))+":"+ss)
	w.Flush()
	return
}

func recvMessage(r *bufio.Reader, message interface{}) {
	length, err := r.ReadString(':')
	if err != nil {
		return
	}

	n, err := strconv.Atoi(length[0 : len(length)-1])
	if err != nil {
		return
	}

	bytes := make([]byte, n, n)
	_, err = io.ReadFull(r, bytes)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, message)
	if err != nil {
		log.Fatal("Can't receive message:", err)
	}
}

func getTournamentScore(punter int, scores []Score) int {
	var myScore int
	for _, sc := range scores {
		if sc.Punter == punter {
			myScore = sc.Score
		}
	}
	rank := 0
	for _, sc := range scores {
		if sc.Punter != punter && sc.Score > myScore {
			rank++
		}
	}
	return len(scores) - rank
}

func handshake(r *bufio.Reader, w *bufio.Writer) {
	me := Me{Me: name}
	sendMessage(w, me)

	var you You
	recvMessage(r, &you)

	if me.Me != you.You {
		log.Fatal("Handshake failed: expected:", me.Me, " received:", you.You)
	}
}

func interact(r *bufio.Reader, w *bufio.Writer) {
	handshake(r, w)

	var step Step
	recvMessage(r, &step)
	if step.Map != nil {
		gm := toGameMap(step.Map)

		var p game.Player
		p.Setup(*step.Punter, *step.Punters, gm)
		log.Println("Punter id:", *step.Punter)
		log.Println("Number of punters:", *step.Punters)
		log.Println("Game map:", gm)

		sendMessage(w, Ready{Ready: p.Punter, State: &p})

		return
	}

	if step.Moves != nil {
		p := step.State
		move := p.MakeMove(toGameMoves(step.Moves.Moves))
		log.Printf("%T Making move: %v", move, move)
		sendMessage(w, fromGameMove(&move, p))
		return
	}

	if step.Stop != nil {
		log.Println("Final scores:", step.Stop.Scores)
		log.Printf("Tournament score: %d/%d\n", getTournamentScore(step.State.Punter, step.Stop.Scores), len(step.Stop.Scores))
		return
	}

	log.Fatal("Unknown state")
}

func main() {
	log.SetFlags(0)

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	interact(reader, writer)
}
