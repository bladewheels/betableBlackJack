package main

// The API defines the structure of the returned message from an API call
type API struct {
	Message               string `json:"message"`
	SomethingElseEntirely string `json:"wtf"`
}

type respStruct struct {
	Args          map[string]string
	Headers       map[string]string
	Origin        string
	Url           string
	Authorization string
}

// Game holds the state of a game of BlackJack
type Game struct {
	GameID    string       `json:"game_id"`
	Deck      Deck         `json:"deck"`
	ShuffleAt int          `json:"shuffle_at"`
	Dealer    PublicDealer `json:"dealer"`
	Player    Player       `json:"player"`
	Winner    string       `json:"winner"`
}

// Player holds the state of a player of a game of BlackJack
type Player struct {
	Cards      []Card `json:"cards"`
	HandTotals []int  `json:"hand_totals"`
}

// Dealer holds the state of a dealer of a game of BlackJack
type Dealer struct {
	Cards      []Card `json:"cards"`
	SecretCard Card   `json:"secret_card,omitempty"`
	HandTotals []int  `json:"hand_totals,omitempty"`
}

type omit *struct{}

// PublicDealer masks the face-down card of the Dealer from JSON serialization
type PublicDealer struct {
	*Dealer
	SecretCard omit `json:"secret_card,omitempty"`
	HandTotals omit `json:"hand_totals,omitempty"`
}

// Deck holds the state of the source of Cards to be drawn upon by the Dealer
type Deck struct {
	DeckID    string `json:"deck_id"`
	Remaining int    `json:"remaining"`
}

// DECK OF CARDS' (https://deckofcardsapi.com) JSON responses are mapped to Go structs below:

// Deck defines the response to a query like: https://deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1
//type Deck struct {
//	DeckID    string `json:"deck_id"`
//	Success   bool   `json:"success"`
//	Remaining int    `json:"remaining"`
//	Shuffled  bool   `json:"shuffled"`
//}

// DeckWithDrawnCards represents...
type DeckWithDrawnCards struct {
	DeckID    string `json:"deck_id"`
	Success   bool   `json:"success"`
	Remaining int    `json:"remaining"`
	Cards     []Card `json:"cards"`
}

// CardDraw represents the result of drawing a number of Cards from a Deck
type CardDraw struct {
	Success   bool   `json:"success"`
	Cards     []Card `json:"cards"`
	DeckID    string `json:"deck_id"`
	Remaining int    `json:"remaining"`
}

// Card represents a card from a deck
type Card struct {
	Image string `json:"image"`
	Value string `json:"value"`
	Suit  string `json:"suit"`
	Code  string `json:"code"`
}

// END of DECK OF CARDS' Go structs
