package main

// Deck defines the response to a query like: https://deckofcardsapi.com/api/deck/new/shuffle/?deck_count=1
type Deck struct {
	DeckID string `json:"deck_id"`
	Success bool `json:"success"`
	Remaining int `json:"remaining"`
	Shuffled bool `json:"shuffled"`
}
// Card represents a card from a deck
type Card struct {
	Suit   string `json:"suit"`
	Image  string `json:"image"`
	Images struct {
		Svg string `json:"svg"`
		Png string `json:"png"`
	} `json:"images"`
	Code  string `json:"code"`
	Value string `json:"value"`
}

// DrawnCards defines the response to a query like: https://deckofcardsapi.com/api/deck/medszzrfkqua/draw/?count=2
type DrawnCards struct {
	DeckID    string `json:"deck_id"`
	Success   bool   `json:"success"`
	Remaining int    `json:"remaining"`
	Cards     []Card `json:"cards"`
}

// TODO: May NOT need this one i.e. just put the hands into a local in-memory struct until we scale beyond 1 player!
// Hand defines the response to a query like: https://deckofcardsapi.com/api/deck/medszzrfkqua/pile/doug/add/?cards=AD
type Hand struct {
	DeckID string `json:"deck_id"`
	Piles struct {
		Player struct {
			Remaining int `json:"remaining"`
		} `json:"player"`
	} `json:"piles"`
	Success bool `json:"success"`
	Remaining int `json:"remaining"`
}

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
