package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 5 * time.Second,
}

type Player struct {
	Conn  *websocket.Conn
	ID    int
	Picks []string
}

type Game struct {
	Players     [2]Player
	PlayerIncId int
	CurrentTurn int
	Round       int
}

var currentGame *Game = &Game{
	Players:     [2]Player{},
	PlayerIncId: 0,
	CurrentTurn: 0,
	Round:       1,
}

var mu *sync.Mutex = &sync.Mutex{}

var footballTeams = loadTeams()

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("Server starting on port :8080")
	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	if currentGame.PlayerIncId >= 2 {
		// Game is full
		mu.Unlock()
		fmt.Printf("Game is full. Can't connect\n")
		w.WriteHeader((http.StatusServiceUnavailable))
		return
	}
	mu.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failure to upgrade HTTP to websocket")
		return
	}

	defer conn.Close()

	mu.Lock()
	player := Player{
		Conn:  conn,
		ID:    currentGame.PlayerIncId,
		Picks: nil,
	}

	fmt.Printf("Player %d joined\n", player.ID+1)

	currentGame.Players[currentGame.PlayerIncId] = player
	currentGame.PlayerIncId += 1
	mu.Unlock()

	msg := "Welcome to the Football draft. Waiting on both participants to join"
	conn.WriteMessage(websocket.TextMessage, []byte(msg))
	for {
		mu.Lock()
		ready := currentGame.PlayerIncId >= 2
		mu.Unlock()
		if ready {
			fmt.Printf("Game is now full\n")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	msg = "Let's Begin!"
	conn.WriteMessage(websocket.TextMessage, []byte(msg))

	for {
		mu.Lock()
		round := currentGame.Round
		if round >= 7 {
			mu.Unlock()
			conn.WriteMessage(websocket.TextMessage, []byte("Draft complete! Game over."))
			fmt.Printf("Rounds have ended\n")
			break
		}

		isMyTurn := currentGame.CurrentTurn == player.ID
		mu.Unlock()

		if isMyTurn {
			fmt.Printf("It is Player %d's turn in round: %d\n", player.ID, round)
			randomTeam := getRandomTeam()
			fmt.Printf("The random team selected was: %s\n", randomTeam)
			msg := fmt.Sprintf("It is round: %d and it is your turn. your randomized team is: %s. Who do you choose?",
				round,
				randomTeam)
			conn.WriteMessage(websocket.TextMessage, []byte(msg))

			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failure reading in message\n", err)
				break
			}

			response := strings.TrimSpace(string(p))
			fmt.Printf("Player %d picked: %s (Round %d)\n", player.ID, response, round)

			mu.Lock()
			currentGame.Players[player.ID].Picks = append(currentGame.Players[player.ID].Picks, response)
			currentGame.CurrentTurn = (currentGame.CurrentTurn + 1) % 2
			if currentGame.CurrentTurn == 0 {
				currentGame.Round++
			}
			mu.Unlock()
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func loadTeams() []string {
	data, err := os.ReadFile("nfl_teams.txt")
	if err != nil {
		log.Fatal("Failure to read in NFL teams")
	}

	teams := strings.Split(strings.TrimSpace(string(data)), "\n")
	return teams
}

func getRandomTeam() string {
	randomIndex := rand.Intn(len(footballTeams))
	return footballTeams[randomIndex]
}
