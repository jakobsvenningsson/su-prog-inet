// Package drawserver ...
package drawserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// DrawServer represents a drawing server.
type DrawServer struct {
	port  string
	peers []string
	lines []line
	ws    *websocket.Conn
}

type point struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type line struct {
	Coordiantes []point `json:"cords"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// New returns a new struct DrawServer.
func New(port string, peers []string) *DrawServer {
	return &DrawServer{
		port:  port,
		peers: peers,
		lines: make([]line, 0),
	}
}

// Listen configures and serves the web-based drawing UI.
func (ds *DrawServer) Listen() error {
	go ds.listenUDP()
	http.HandleFunc("/", ds.canvas)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/ws", ds.newWs)
	http.HandleFunc("/lines", ds.getLines)
	return http.ListenAndServe(fmt.Sprintf(":%s", ds.port), nil)
}

func (ds *DrawServer) listenUDP() {
	port, err := strconv.Atoi(ds.port)
	if err != nil {
		log.Fatal(err)
	}

	// Listen for incoming UDP packets
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 8192)
	for {
		// Read lines sent from peers.
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		// Decode and save line.
		var coordinates []point
		if err = json.NewDecoder(bytes.NewBuffer(buf[:n])).Decode(&coordinates); err != nil {
			log.Fatal(err)
		}
		ds.lines = append(ds.lines, line{Coordiantes: coordinates})

		// If client websocket connection exists, then send notification.
		if ds.ws == nil {
			continue
		}
		if err = ds.ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			log.Fatal(err)
		}
	}
}

func (ds *DrawServer) newWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	ds.ws = ws
	if err != nil {
		log.Println(err)
	}
	for {
		// read in a message
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var coordinates []point
		if err = json.NewDecoder(bytes.NewBuffer(p)).Decode(&coordinates); err != nil {
			log.Fatal(err)
		}
		ds.lines = append(ds.lines, line{Coordiantes: coordinates})

		for _, peerPort := range ds.peers {
			addr := fmt.Sprintf("127.0.0.1:%s", peerPort)
			fmt.Println("sending to addr")
			conn, err := net.Dial("udp", addr)
			if err != nil {
				panic(err)
			}
			defer conn.Close()
			_, err = conn.Write(p)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (ds *DrawServer) canvas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.ParseFiles("templates/main.html")
	if err != nil {
		log.Fatal(err)
	}
	data := struct {
		Port string
	}{
		Port: ds.port,
	}
	t.Execute(w, data)
}

func (ds *DrawServer) getLines(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ds.lines)
}
