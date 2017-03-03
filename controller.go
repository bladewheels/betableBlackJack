package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var gameChannelMap = make(map[string]chan Game)

// Pull a parameter out of the HTTP Request
func getParamFromRequest(param string, r *http.Request) string {
	return mux.Vars(r)[param]
}

// The response handler for requests to the /api/games URI; TODO: limit to POST only
func start(w http.ResponseWriter, r *http.Request) {

	game, err := getGameStarted()
	if err != nil {
		fmt.Println("Failed to get Game!")
		fmt.Fprintf(w, "Failed to get Game, please try again later.")
		return
	}

	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// The response handler for requests to the /api/games/{gameID}/hit URI; TODO: limit to PUT only
func hit(w http.ResponseWriter, r *http.Request) {

	game, err := hitPlayer(getParamFromRequest("gameID", r))
	if err != nil {
		fmt.Println("Failed to hit Player!")
		fmt.Fprintf(w, "Failed to hit Player, please try again.")
		return
	}

	// Prepare the return value
	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// The response handler for requests to the /api/game/gameId/stand URI; TODO: limit to PUT only
func stand(w http.ResponseWriter, r *http.Request) {

	game, err := playForDealer(getParamFromRequest("gameID", r))
	if err != nil {
		if strings.Contains(err.Error(), "dequeue") {
			fmt.Fprintf(w, "You cannot stand on a completed Game")
			return
		}
		if strings.Contains(err.Error(), "draw") {
			fmt.Fprintf(w, "Failed to draw a card, please try again")
			return
		}
	}

	// Prepare the return value
	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}
	deRegisterGame(game.GameID)

	fmt.Fprintf(w, string(output))
}

func main() {
	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/api/games", start)
	gRouter.HandleFunc("/api/games/{gameID}/hit", hit)
	gRouter.HandleFunc("/api/games/{gameID}/stand", stand)

	http.Handle("/", gRouter)
	http.ListenAndServe(":8080", nil)
}
