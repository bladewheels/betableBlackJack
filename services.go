package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Hit the player
func hitPlayer(gameID string) (Game, error) {
	var g Game

	// Grab the Game; it may be in-progress or completed i.e. Player BUST
	game, err := dequeueGame(gameID)
	if err != nil {
		msg := "Failed to get the Game! (i.e. #: " + gameID + ")"
		fmt.Println(msg)
		return g, errors.New(msg)
	}

	// Draw a card from the Deck and add it to the Player's hand
	newCard, err := getNewCard(gameID)
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
		revealDealerHand(game)
		deRegisterGame(game.GameID)
	}
	return game, nil
}

// Start the game i.e. get a deck, deal some cards and deal with naturals, if any
func getGameStarted() (Game, error) {
	var g Game

	// Get a Deck of Cards with 4 Cards already drawn from the Deck
	deckCount := 6
	cardCount := 4
	deckWithCards, err := getNewShuffledDeckWithCards(deckCount, cardCount)
	if err != nil {
		fmt.Println("Failed to get Deck!")
		return g, errors.New("Failed to get Deck!")
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
	return game, nil
}

func playForDealer(gameID string) (Game, error) {

	// Grab the Game; it may be in-progress or completed i.e. Player BUST
	game, err := dequeueGame(gameID)
	if err != nil {
		return game, errors.New("Failed to dequeue the Game #" + gameID + "!")
	}

	for dealerShouldHit(game) {

		newCard, err2 := getNewCard(game.GameID)
		if err2 != nil {
			fmt.Println("Something went wrong drawing a card!")
		}
		game.Dealer.Dealer.Cards = append(game.Dealer.Dealer.Cards, newCard)

		// Calculate the Dealer's hand total(s)
		// TODO: Refactor this block into a function, it is used elsewhere
		dealerHandTotals := make([]int, 1)
		dealerHandTotals, _ = updateHandTotal(dealerHandTotals, game.Dealer.Dealer.SecretCard)
		for _, card := range game.Dealer.Dealer.Cards {
			dealerHandTotals, _ = updateHandTotal(dealerHandTotals, card)
		}
		game.Dealer.Dealer.HandTotals = dealerHandTotals
	}
	game.Winner = determineWinnerAtEndOfGame(game)
	revealDealerHand(game)
	return game, nil
}

// Get a new Card; retry up to 3 times if at first you don't succeed...
func getNewCard(deckID string) (Card, error) {
	var newCard Card
	err := Do(func(attempt int) (bool, error) {
		var err error
		card, err := getCardFromDeck(deckID)
		if err != nil {
			time.Sleep(200 * time.Millisecond) // wait a sec
		}
		newCard = card
		return attempt < 3, err // try 3 times
	})
	if err != nil {
		return newCard, errors.New("Failed to draw a card!")
	}
	return newCard, nil
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

	if len(cardDraw.Cards) < 1 {
		// TODO: retry 3 times
		return c, errors.New("The CardDraw did NOT have a card in it!")
	}

	return cardDraw.Cards[0], nil
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

// Reveal the Dealer hand for endgame display, put the secret card first
func revealDealerHand(game Game) {
	var allCardsWithSecretCardFirst []Card
	allCardsWithSecretCardFirst = append(allCardsWithSecretCardFirst, game.Dealer.Dealer.SecretCard)
	for _, card := range game.Dealer.Dealer.Cards {
		allCardsWithSecretCardFirst = append(allCardsWithSecretCardFirst, card)
	}
	game.Dealer.Dealer.Cards = allCardsWithSecretCardFirst
}

// Determine if the Dealer should hits
func dealerShouldHit(game Game) bool {
	/*
	   TODO: Verify/implement the following requirement:
	   i.e. If the dealer has an ace, and counting it as 11 would bring his total to 17 or more (but not over 21),
	   he must count the ace as 11 and stand - See more at: http://www.bicyclecards.com/how-to-play/blackjack/
	*/
	bestPlayerTotal := maxHandTotalUnderLimit(game.Player.HandTotals)
	bestDealerTotal := maxHandTotalUnderLimit(game.Dealer.Dealer.HandTotals)

	if bestPlayerTotal == 21 {
		if bestPlayerTotal == bestDealerTotal {
			return false
		} else if bestPlayerTotal > bestDealerTotal {
			return true
		}
	} else if bestDealerTotal >= 17 {
		return false
	} else if bestDealerTotal == 0 { // i.e. Dealer BUST
		return false
	}
	return true
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

// Remove the channel that carries the Game state
func deRegisterGame(gameID string) {
	delete(gameChannelMap, gameID)
}

// Determine a winner, if any, after a Player is hit
func determineWinnerAfterPlayerHit(game Game) string {
	bestPlayerTotal := maxHandTotalUnderLimit(game.Player.HandTotals)

	if bestPlayerTotal == 0 { // i.e. BUST
		return "dealer"
	}
	return "none"
}

// Get the maximum hand total that does not exceed 21, returns zero if BUST
func maxHandTotalUnderLimit(handTotals []int) int {
	max := 0
	for _, total := range handTotals {
		if total <= 21 && total > max {
			max = total
		}
	}
	return max
}

// Determine a winner, if any, after a Dealer is hit
func determineWinnerAfterDealerHit(game Game) string {
	bestDealerTotal := maxHandTotalUnderLimit(game.Dealer.Dealer.HandTotals)

	if bestDealerTotal == 0 { // i.e. BUST
		return "player"
	}
	return "none"
}

// Determine the winner of a game, if any, after the Dealer stands
func determineWinnerAtEndOfGame(game Game) string {
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
	} else if bestPlayerTotal == bestDealerTotal {
		return "both"
	} else if bestPlayerTotal > bestDealerTotal {
		return "player"
	}
	return "dealer"
}

// Determine the winner of a game, if any, after the initial deal
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
}
