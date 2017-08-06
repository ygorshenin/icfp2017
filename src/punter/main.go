package main

import (
	"bufio"
	"common"
	"encoding/json"
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

type Ready struct {
	Ready int                 `json:"ready"`
	State *common.PlayerProxy `json:"state"`
}

type Score struct {
	Punter int `json:"punter"`
	Score  int `json:"score"`
}

type Stop struct {
	Moves  []common.Move `json:"move"`
	Scores []Score       `json:"scores"`
}

type Step struct {
	Punter  *int        `json:"punter"`
	Punters *int        `json:"punters"`
	Map     *common.Map `json:"map"`

	Moves *common.Moves       `json:"move"`
	Stop  *Stop               `json:"stop"`
	State *common.PlayerProxy `json:"state"`
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
		log.Fatal("Can't receive message:", err, " [", string(bytes), "]")
	}
}

func formatScores(punter int, scores []Score) string {
	s := "["
	for _, sc := range scores {
		if len(s) > 1 {
			s += ", "
		}
		if sc.Punter == punter {
			s += "[" + strconv.Itoa(sc.Score) + "]"
		} else {
			s += strconv.Itoa(sc.Score)
		}
	}
	s += "]"
	return s
}

func getRank(punter int, scores []Score) int {
	var myScore int
	for _, sc := range scores {
		if sc.Punter == punter {
			myScore = sc.Score
		}
	}
	rank := 1
	for _, sc := range scores {
		if sc.Punter != punter && sc.Score > myScore {
			rank++
		}
	}
	return rank
}

func handshake(r *bufio.Reader, w *bufio.Writer, n string) {
	me := Me{Me: name + ": " + n}
	sendMessage(w, me)

	var you You
	recvMessage(r, &you)

	if me.Me != you.You {
		log.Fatal("Handshake failed: expected:", me.Me, " received:", you.You)
	}
}

func interact(r *bufio.Reader, w *bufio.Writer) {
	pp := common.MakePlayerProxy("baseline")
	handshake(r, w, pp.Name())

	var step Step
	step.State = &pp

	recvMessage(r, &step)
	if step.Map != nil {
		pp.Setup(*step.Punter, *step.Punters, step.Map)
		log.Println("Punter id:", *step.Punter)
		log.Println("Number of punters:", *step.Punters)
		log.Println("Game map:", *step.Map)

		sendMessage(w, Ready{Ready: *step.Punter, State: &pp})

		return
	}

	if step.Moves != nil {
		move := pp.MakeMove(step.Moves.Moves)
		log.Printf("Making move: %v", move.String())
		sendMessage(w, move)
		return
	}

	if step.Stop != nil {
		punter := step.State.GetPunter()
		log.Println("Final scores:", formatScores(punter, step.Stop.Scores))
		log.Printf("Rank: %d/%d\n", getRank(punter, step.Stop.Scores), len(step.Stop.Scores))
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
