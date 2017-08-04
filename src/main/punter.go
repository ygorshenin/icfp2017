package main

import (
	"bufio"
	"encoding/json"
	"flag"
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

func sendMessage(w *bufio.Writer, message interface{}) {
	bs, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	ss := string(bs)

	io.WriteString(w, strconv.Itoa(len(ss))+":"+ss)
	w.Flush()
}

func recvMessage(r *bufio.Reader, message interface{}) (err error) {
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

	return json.Unmarshal(bytes, message)
}

func handshake(r *bufio.Reader, w *bufio.Writer) {
	me := Me{Me: name}
	sendMessage(w, me)

	var you You
	err := recvMessage(r, &you)
	if err != nil {
		log.Fatal("Can't do hanshake with server:", err)
	}

	if me.Me != you.You {
		log.Fatal("Expected:", me.Me, "received:", you.You)
	}

	log.Println("Successful handshake")
}

func main() {
	log.SetFlags(0)
	flag.Parse()

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
	}
}
