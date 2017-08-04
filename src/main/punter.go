package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"game"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	name      = "lambda"
	serverUrl = "punter.inf.ed.ac.uk"
)

var flagPort = flag.Int("port", -1, "port for online mode, negative value means offline mode")

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

type Setup struct {
	Punter  int `json:"punter"`
	Punters int `json:"punters"`
	Map     Map `json:"map"`
}

type Ready struct {
	Ready int `json:"ready"`
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
	Claim *ClaimMove `json:"claim,omitempty"`
	Pass  *PassMove  `json:"pass,omitempty"`
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

func fromGameMove(m *game.Move) (r Move) {
	switch m.Type {
	case game.Claim:
		r.Claim = &ClaimMove{Punter: m.Punter, Source: m.Source, Target: m.Target}
	case game.Pass:
		r.Pass = &PassMove{Punter: m.Punter}
	default:
		log.Fatal("Unknown move type:", m.Type)
	}
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
	Moves   *Moves `json:"move"`
	Stop    *Stop  `json:"stop"`
	Timeout *int   `json:"timeout"`
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

func handshake(r *bufio.Reader, w *bufio.Writer) {
	me := Me{Me: name}
	sendMessage(w, me)

	var you You
	recvMessage(r, &you)

	if me.Me != you.You {
		log.Fatal("Expected:", me.Me, " received:", you.You)
	}

	log.Println("Successful handshake")
}

func setup(r *bufio.Reader, w *bufio.Writer, p *game.Player) {
	var setup Setup
	recvMessage(r, &setup)

	gm := toGameMap(&setup.Map)

	log.Println("Punter id:", setup.Punter)
	log.Println("Number of punters:", setup.Punters)
	log.Println("Game map:", gm)

	p.Setup(setup.Punter, setup.Punters, gm)

	sendMessage(w, Ready{Ready: setup.Punter})
}

func interact(r *bufio.Reader, w *bufio.Writer, p *game.Player) {
	for {
		var step Step
		recvMessage(r, &step)
		if step.Moves != nil {
			move := p.MakeMove(toGameMoves(step.Moves.Moves))
			log.Println("Making move:", move)
			sendMessage(w, fromGameMove(&move))
			continue
		}
		if step.Stop != nil {
			log.Println("Final scores:", step.Stop.Scores)
			break
		}
		if step.Timeout != nil {
			log.Println("Timeout:", *step.Timeout)
			continue
		}
		log.Fatal("Unknown state")
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	var player game.Player

	if *flagPort < 0 {
		log.Println("Running in offline mode")
	} else {
		log.Println("Running in online mode")

		conn, err := net.Dial("tcp", serverUrl+":"+strconv.Itoa(*flagPort))
		if err != nil {
			log.Fatal("Can't dial connection:", err)
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		handshake(reader, writer)
		setup(reader, writer, &player)
		interact(reader, writer, &player)
	}
}
