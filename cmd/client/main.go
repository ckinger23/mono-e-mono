package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ckinger23/mono-e-mono/internal/models"
	"github.com/ckinger23/mono-e-mono/internal/protocol"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn         *websocket.Conn
	playerNumber int
	roster       []models.RosterSlot
	done         chan struct{}
}

func NewClient(serverAddr string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/ws"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn: conn,
		done: make(chan struct{}),
	}, nil
}

func (c *Client) Run() {
	defer c.conn.Close()

	// Start message reader
	go c.readMessages()

	// Wait for game to end
	<-c.done
	fmt.Println("\nThanks for playing!")
}

func (c *Client) readMessages() {
	defer close(c.done)

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Connection error: %v", err)
			}
			return
		}

		var msg protocol.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) handleMessage(msg *protocol.Message) {
	switch msg.Type {
	case protocol.MsgWelcome:
		c.handleWelcome(msg)
	case protocol.MsgWaitingPlayer:
		c.handleWaiting(msg)
	case protocol.MsgGameStart:
		c.handleGameStart(msg)
	case protocol.MsgYourTurn:
		c.handleYourTurn(msg)
	case protocol.MsgWaitTurn:
		c.handleWaitTurn(msg)
	case protocol.MsgPickConfirmed:
		c.handlePickConfirmed(msg)
	case protocol.MsgOpponentPicked:
		c.handleOpponentPicked(msg)
	case protocol.MsgGameOver:
		c.handleGameOver(msg)
	case protocol.MsgError:
		c.handleError(msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (c *Client) handleWelcome(msg *protocol.Message) {
	payload := parsePayload[protocol.WelcomePayload](msg.Payload)

	c.playerNumber = payload.PlayerNumber

	clearScreen()
	printHeader()
	fmt.Printf("\n%s\n", payload.Message)
	fmt.Printf("You are Player %d\n", payload.PlayerNumber)
}

func (c *Client) handleWaiting(msg *protocol.Message) {
	payload := parsePayload[map[string]string](msg.Payload)
	fmt.Printf("\n%s\n", payload["message"])
}

func (c *Client) handleGameStart(msg *protocol.Message) {
	payload := parsePayload[protocol.GameStartPayload](msg.Payload)

	clearScreen()
	printHeader()
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("  %s\n", payload.Message)
	fmt.Printf("  You are playing against: %s\n", payload.OpponentLabel)
	fmt.Println(strings.Repeat("=", 50))
}

func (c *Client) handleYourTurn(msg *protocol.Message) {
	payload := parsePayload[protocol.YourTurnPayload](msg.Payload)

	c.roster = payload.YourRoster

	clearScreen()
	printHeader()
	fmt.Println()
	fmt.Printf("  ROUND %d - YOUR TURN\n", payload.Round)
	fmt.Println(strings.Repeat("-", 50))

	// Show current roster
	fmt.Println("\n  YOUR ROSTER:")
	printRoster(payload.YourRoster)

	// Show the drawn team
	fmt.Printf("\n  Random Team Drawn: %s (%s)\n", payload.TeamName, payload.TeamAbbrev)
	fmt.Println(strings.Repeat("-", 50))

	// Show available players
	fmt.Println("\n  AVAILABLE PLAYERS:")
	if len(payload.AvailablePlayers) == 0 {
		fmt.Println("  No players available that fit your open roster slots!")
		fmt.Println("  The team will be re-drawn...")
		return
	}

	for i, player := range payload.AvailablePlayers {
		fmt.Printf("    %d. %-25s (%s)\n", i+1, player.Name, player.Position)
	}

	// Get player selection
	fmt.Println()
	fmt.Print("  Enter the number of the player you want to draft: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(payload.AvailablePlayers) {
		fmt.Println("  Invalid selection. Please enter a valid number.")
		return
	}

	selectedPlayer := payload.AvailablePlayers[selection-1]

	// Send the pick
	c.sendPick(selectedPlayer.ID, selectedPlayer.Position)
}

func (c *Client) handleWaitTurn(msg *protocol.Message) {
	payload := parsePayload[protocol.WaitTurnPayload](msg.Payload)

	clearScreen()
	printHeader()
	fmt.Println()
	fmt.Printf("  ROUND %d\n", payload.Round)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()
	fmt.Printf("  %s\n", payload.Message)
	fmt.Println()

	// Show current roster
	if len(c.roster) > 0 {
		fmt.Println("  YOUR ROSTER:")
		printRoster(c.roster)
	}
}

func (c *Client) handlePickConfirmed(msg *protocol.Message) {
	payload := parsePayload[protocol.PickConfirmedPayload](msg.Payload)

	fmt.Println()
	fmt.Printf("  Pick confirmed: %s (%s)\n", payload.PlayerName, payload.Position)
}

func (c *Client) handleOpponentPicked(msg *protocol.Message) {
	payload := parsePayload[protocol.OpponentPickedPayload](msg.Payload)

	fmt.Println()
	fmt.Printf("  Player %d picked %s (%s) from %s\n",
		payload.OpponentNumber, payload.PlayerName, payload.Position, payload.TeamName)
}

func (c *Client) handleGameOver(msg *protocol.Message) {
	payload := parsePayload[protocol.GameOverPayload](msg.Payload)

	clearScreen()
	printHeader()
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("  DRAFT COMPLETE!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Show final rosters
	fmt.Println("  YOUR FINAL ROSTER:")
	printRoster(payload.YourRoster)
	fmt.Printf("  Your Score: %.1f points\n", payload.YourScore)

	fmt.Println()
	fmt.Println("  OPPONENT'S FINAL ROSTER:")
	printRoster(payload.OpponentRoster)
	fmt.Printf("  Opponent Score: %.1f points\n", payload.OpponentScore)

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))

	// Show result
	if payload.Winner == 0 {
		fmt.Println("  IT'S A TIE!")
	} else if payload.Winner == c.playerNumber {
		fmt.Println("  CONGRATULATIONS! YOU WON!")
	} else {
		fmt.Println("  YOU LOST. BETTER LUCK NEXT TIME!")
	}

	fmt.Printf("  Final Score: %.1f - %.1f\n", payload.YourScore, payload.OpponentScore)
	fmt.Println(strings.Repeat("=", 50))
}

func (c *Client) handleError(msg *protocol.Message) {
	payload := parsePayload[protocol.ErrorPayload](msg.Payload)
	fmt.Printf("\n  ERROR: %s\n", payload.Message)
}

func (c *Client) sendPick(playerID string, position models.Position) {
	msg := protocol.Message{
		Type: protocol.MsgPickPlayer,
		Payload: protocol.PickPlayerPayload{
			PlayerID: playerID,
			Position: position,
		},
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send pick: %v", err)
	}
}

func parsePayload[T any](payload any) T {
	var result T
	data, _ := json.Marshal(payload)
	json.Unmarshal(data, &result)
	return result
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printHeader() {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("         MONO-E-MONO Fantasy Draft")
	fmt.Println(strings.Repeat("=", 50))
}

func printRoster(slots []models.RosterSlot) {
	for _, slot := range slots {
		if slot.Player != nil {
			fmt.Printf("    %-4s: %s\n", slot.Position, slot.Player.Name)
		} else {
			fmt.Printf("    %-4s: [EMPTY]\n", slot.Position)
		}
	}
}

func main() {
	serverAddr := "localhost:8080"
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	fmt.Printf("Connecting to server at %s...\n", serverAddr)

	client, err := NewClient(serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client.Run()
}
