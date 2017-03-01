package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/bndr/gopencils"
	"github.com/gorilla/mux"
)

// Hello is the default response handler for requests to the /api/userId URI
func Hello(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	name := urlParams["user"]
	HelloMessage := "Hello, " + name

	message := API{HelloMessage, ""}
	output, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Something went wrong!")
	}
	sharpenPencils()
	fmt.Fprintf(w, string(output))

}

// GetGame is the response handler for requests to the /api/game URI
func GetGame(w http.ResponseWriter, r *http.Request) {

	HelloMessage := "Welcome to BlackJack!"
	SomethingElseEntirely := "gameOn"

	message := API{HelloMessage, SomethingElseEntirely}
	output, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// HitMe is the response handler for requests to the /api/game/gameId/hit URI
func HitMe(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	game := urlParams["game"]

	HitMeMessage := "Here you Go! (for game #" + game + ")"
	SomethingElseEntirely := "aNewCardOfSomeType"

	message := API{HitMeMessage, SomethingElseEntirely}
	output, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// Stand is the response handler for requests to the /api/game/gameId/stand URI
func Stand(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	game := urlParams["game"]

	HitMeMessage := "OK, I'll see if I can beat you! (for game #" + game + ")"
	SomethingElseEntirely := "Good Luck!"

	message := API{HitMeMessage, SomethingElseEntirely}
	output, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

func sharpenPencils() {
    api := gopencils.Api("https://api.github.com")
    // Users Resource
    users := api.Res("users")

    usernames := []string{"bndr", "torvalds", "coleifer"}

    for _, username := range usernames {
        // Create a new pointer to response Struct
        r := new(respStruct)
        // Get user with id i into the newly created response struct
        _, err := users.Id(username, r).Get()
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Println(r)
        }
    }
}

func main() {
	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/api/{user:[0-9]+}", Hello)
	gRouter.HandleFunc("/api/game", GetGame)
	gRouter.HandleFunc("/api/game/{game:[0-9]+}/hit", HitMe)
	gRouter.HandleFunc("/api/game/{game:[0-9]+}/stand", Stand)
	http.Handle("/", gRouter)
	http.ListenAndServe(":8080", nil)
}
