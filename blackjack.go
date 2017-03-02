package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

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
	}

	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}

	fmt.Fprintf(w, string(output))
}

// Determine a winner of a game, if any
func determineWinnerAtStartOfGame(game Game) string {
	bestPlayerTotal := maxHandTotalUnderLimit(game.Player.HandTotals)
	bestDealerTotal := maxHandTotalUnderLimit(game.Dealer.Dealer.HandTotals)

	if bestPlayerTotal == 21 {
		if bestPlayerTotal == bestDealerTotal {
			return "both"
		} else if bestPlayerTotal > bestDealerTotal {
			return "player"
		}
	} else if bestDealerTotal == 21 {
		return "dealer"
	}
	return "none"
	// Iff Player has won then: defer dequeueGame(gameID)
}

//
func determineWinnerAfterPlayerHit(game Game) string {
	bestPlayerTotal := maxHandTotalUnderLimit(game.Player.HandTotals)

	if bestPlayerTotal == 0 { // i.e. BUST
		return "dealer"
	}
	return "none"
}

// Get the maximum hand total that does not exceed 21
func maxHandTotalUnderLimit(handTotals []int) int {
	max := 0
	for _, total := range handTotals {
		if total <= 21 && total > max {
			max = total
		}
	}
	return max
}

// Update the Player or Dealer hand total(s) with the value of a Card
func updateHandTotal(handTotals []int, card Card) ([]int, error) {
	cardValue, err := strconv.ParseInt(card.Value, 10, 0)
	if err != nil {
		// Must be an Ace or face Card
		if card.Value == "ACE" {
			for i, initialTotal := range handTotals {
				handTotals[i] = initialTotal + 1
				handTotals = append(handTotals, initialTotal+11)
			}
		} else {
			for i, initialTotal := range handTotals {
				handTotals[i] = initialTotal + 10
			}
		}
	} else {
		for i, initialTotal := range handTotals {
			handTotals[i] = initialTotal + int(cardValue)
		}
	}
	return handTotals, nil
}

// Queue the Game state in a channel for future use
func queueGame(game Game) {
	channel, ok := gameChannelMap[game.GameID]
	if !ok {
		channel = make(chan Game, 1) // i.e. non-blocking w/capacity of 1
		gameChannelMap[game.GameID] = channel
	}

	channel <- game
}

// Retrieve the Game state from a channel
func dequeueGame(gameID string) (Game, error) {
	var g Game
	channel, ok := gameChannelMap[gameID]
	if !ok {
		return g, errors.New("Failed to dequeue the Game#" + gameID + "!")
	}

	return <-channel, nil
}

// Remove the channel that carries the Game state
func deleteGame(gameID string) {
	delete(gameChannelMap, gameID)
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

// Return a new, shuffled deck of playing cards w/some cards already drawn
func getNewShuffledDeckWithCards(deckCount, cardCount int) (*DeckWithDrawnCards, error) {

	res, err := http.Get("https://deckofcardsapi.com/api/deck/new/draw/?count=" + strconv.Itoa(cardCount) + "&deck_count=" + strconv.Itoa(deckCount))
	if err != nil {
		fmt.Println("Something went wrong!")
		return nil, err
	}

	body, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		fmt.Println("Something went wrong!")
		return nil, err2
	}

	deck := new(DeckWithDrawnCards)
	err3 := json.Unmarshal(body, &deck)
	if err3 != nil {
		fmt.Println("Something went wrong!")
		return nil, err3
	}

	return deck, nil
}

// Return a playing card from the deck
func getCardFromDeck(deckID string) (Card, error) {

	var c Card

	res, err := http.Get("https://deckofcardsapi.com/api/deck/" + deckID + "/draw/?count=1")
	if err != nil {
		fmt.Println("Failed to get a Card from the (remote) Deck!")
		return c, err
	}

	body, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		fmt.Println("Failed to read the body of the response!")
		return c, err2
	}

	cardDraw := new(CardDraw)
	err3 := json.Unmarshal(body, &cardDraw)
	if err3 != nil {
		fmt.Println("Failed to unmarshal the CardDraw!")
		return c, err3
	}

	return cardDraw.Cards[0], nil
}

func main() {
	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/api/games", startGame)
	gRouter.HandleFunc("/api/games/{gameID}/hit", hitPlayer)
	gRouter.HandleFunc("/api/games/{gameID}/stand", playerStands)

	http.Handle("/", gRouter)
	http.ListenAndServe(":8080", nil)
}
