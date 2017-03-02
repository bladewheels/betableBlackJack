package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var gameChannelMap = make(map[string]chan Game)

// The response handler for requests to the /api/games URI; TODO: limit to POST only
func startGame(w http.ResponseWriter, r *http.Request) {

	// Get a Deck of Cards with 4 Cards already drawn from the Deck
	deckCount := 6
	cardCount := 4
	deckWithCards, err := getNewShuffledDeckWithCards(deckCount, cardCount)
	if err != nil {
		fmt.Println("Failed to get Deck!")
		fmt.Fprintf(w, "Failed to get Deck, please try again later.")
		return
	}

	// Deal hands to Player and Dealer
	player := &Player{}
	player.Cards = append(player.Cards, deckWithCards.Cards[0])
	dealer := &Dealer{}
	dealer.Cards = append(dealer.Cards, deckWithCards.Cards[1])
	player.Cards = append(player.Cards, deckWithCards.Cards[2])
	dealer.SecretCard = deckWithCards.Cards[3]

	// Calculate point totals for hands...
	// ...of the Player...
	playerHandTotals := make([]int, 1)
	for _, card := range player.Cards {
		playerHandTotals, _ = updateHandTotal(playerHandTotals, card)
	}
	player.HandTotals = playerHandTotals

	// ...and the Dealer
	dealerHandTotals := make([]int, 1)
	dealerHandTotals, _ = updateHandTotal(dealerHandTotals, dealer.SecretCard)
	for _, card := range dealer.Cards {
		dealerHandTotals, _ = updateHandTotal(dealerHandTotals, card)
	}
	dealer.HandTotals = dealerHandTotals

	// Prepare the return value i.e. the Game state
	deck := &Deck{DeckID: deckWithCards.DeckID, Remaining: deckWithCards.Remaining}
	game := Game{deck.DeckID, *deck, 52*deckCount - 75, PublicDealer{Dealer: dealer}, *player, ""}
	game.Winner = determineWinnerAtStartOfGame(game)
	if game.Winner == "none" {
		defer queueGame(game) // ready to receive hits
	} else {
		revealDealerHand(game)
	}

	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// The response handler for requests to the /api/games/{gameID}/hit URI; TODO: limit to POST only
func hitPlayer(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	gameID := urlParams["gameID"]

	// Grab the Game; it may be in-progress or completed i.e. Player BUST
	game, err := dequeueGame(gameID)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "You cannot hit on a completed Game")
		return
	}

	// Draw a card from the Deck and add it to the Player's hand
	newCard, err := getCardFromDeck(gameID)
	if err != nil {
		fmt.Println("Something went wrong drawing a card!")
	}
	game.Player.Cards = append(game.Player.Cards, newCard)

	// Calculate the Player's hand total(s)
	playerHandTotals := make([]int, 1)
	for _, card := range game.Player.Cards {
		playerHandTotals, _ = updateHandTotal(playerHandTotals, card)
	}
	game.Player.HandTotals = playerHandTotals

	// Determine if the Player has gone BUST i.e. the Dealer has won
	game.Winner = determineWinnerAfterPlayerHit(game)
	if game.Winner == "none" {
		defer queueGame(game) // ready to receive more hits
	} else if game.Winner == "dealer" {
		deleteGame(game.GameID) // unable to receive more hits
	}

	// Prepare the return value
	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// The response handler for requests to the /api/game/gameId/stand URI
func playerStands(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	gameID := urlParams["gameID"]

	// Grab the Game; it may be in-progress or completed i.e. Player BUST
	game, err := playForDealer(gameID)
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
	defer deleteGame(game.GameID)

	fmt.Fprintf(w, string(output))
}

func main() {
	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/api/games", startGame)
	gRouter.HandleFunc("/api/games/{gameID}/hit", hitPlayer)
	gRouter.HandleFunc("/api/games/{gameID}/stand", playerStands)

	http.Handle("/", gRouter)
	http.ListenAndServe(":8080", nil)
}
