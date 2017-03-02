package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bndr/gopencils"
	"github.com/gorilla/mux"
)

var gameChannelMap = make(map[string]chan Game)

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

// The response handler for requests to the /api/games URI; TODO: limit to POST only
func startGame(w http.ResponseWriter, r *http.Request) {

	deckCount := 6
	cardCount := 4
	deckWithCards, err := getNewShuffledDeckWithCards(deckCount, cardCount)
	if err != nil {
		fmt.Println("Something went wrong!")
	}
	fmt.Println("Got Deck!")

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
	fmt.Println(dealer.HandTotals)

	deck := &Deck{DeckID: deckWithCards.DeckID, Remaining: deckWithCards.Remaining}
	game := Game{deck.DeckID, *deck, 52*deckCount - 75, PublicDealer{Dealer: dealer}, *player}
	defer queueGame(game)

	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}
	fmt.Println("Got Game as JSON! i.e. ", string(output))

	fmt.Fprintf(w, string(output))
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

// The response handler for requests to the /api/games/{gameID}/hit URI; TODO: limit to POST only
func hitMe(w http.ResponseWriter, r *http.Request) {

	urlParams := mux.Vars(r)
	gameID := urlParams["gameID"]

	game, err := dequeueGame(gameID)
	if err != nil {
		fmt.Println("Something went wrong dequeuing the Game#" + gameID + "!")
	}
	//fmt.Println(game)

	newCard, err := getCardFromDeck(gameID)
	if err != nil {
		fmt.Println("Something went wrong drawing a card!")
	}
	fmt.Println(newCard)
	game.Player.Cards = append(game.Player.Cards, newCard)
	fmt.Println(game)
	defer queueGame(game)

	//HitMeMessage := "Here you Go! (for game #" + gameID + ")"
	//SomethingElseEntirely := "aNewCardOfSomeType"

	//message := API{HitMeMessage, SomethingElseEntirely}
	//output, err := json.Marshal(message)

	output, err := json.Marshal(game)
	if err != nil {
		fmt.Println("Something went wrong!")
	}
	fmt.Println("Updated Game with new Card! i.e. ", string(output))

	fmt.Fprintf(w, string(output))
}

// The response handler for requests to the /api/game/gameId/stand URI
func stand(w http.ResponseWriter, r *http.Request) {

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

// a POC test of the gopencils library
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

// Return a new, shuffled deck of playing cards w/some cards already drawn
func getNewShuffledDeckWithCards(deckCount, cardCount int) (*DeckWithDrawnCards, error) {

	fmt.Println("Getting Deck and Cards...")
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
	fmt.Println(*deck)

	fmt.Println("Returning Deck and Cards!")
	return deck, nil
}

// Return a playing card from the deck
func getCardFromDeck(deckID string) (Card, error) {

	var c Card
	fmt.Println("Getting CardDraw...")

	res, err := http.Get("https://deckofcardsapi.com/api/deck/" + deckID + "/draw/?count=1")
	if err != nil {
		fmt.Println("Something went wrong!")
		return c, err
	}

	body, err2 := ioutil.ReadAll(res.Body)
	if err2 != nil {
		fmt.Println("Something went wrong!")
		return c, err2
	}

	cardDraw := new(CardDraw)
	err3 := json.Unmarshal(body, &cardDraw)
	if err3 != nil {
		fmt.Println("Something went wrong!")
		return c, err3
	}

	fmt.Println("Got Cards! From helper fn")
	return cardDraw.Cards[0], nil
}

func main() {
	gRouter := mux.NewRouter()
	//gRouter.HandleFunc("/api/{user:[0-9]+}", Hello)
	gRouter.HandleFunc("/api/games", startGame)
	gRouter.HandleFunc("/api/games/{gameID}/hit", hitMe)
	gRouter.HandleFunc("/api/games/{gameID}/stand", stand)
	http.Handle("/", gRouter)
	http.ListenAndServe(":8080", nil)
}
