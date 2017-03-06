# betableBlackJack

## Demo API in Go for a single-user BlackJack game

The nature of HTTP is call and response so this BlackJack gameplay is most straightforward for a single PLAYER in a GAME; polling (or eg WebSockets, in a future implementation) could be utilized for multi-PLAYER games but this initial design is INTENDED FOR SINGLE PLAYER USE only, due to time-to-market constraints.

Concurrent, single-player games with naturals ARE supported.

Splitting pairs is not currently supported in this implementation.

The current implementation responds to either POST or GET requests for ease of exploration i.e. GETs with a web browser or HTTP tool. It is intended that the API implement only POST/PATCH endpoints so as to remain RESTful.

### Example JSON responses

![The initial deal and after a few hits](/../screenshots/betableBlackJackDemo.drawAndHit.png?raw=true "Start of a game and after a few hits")

### Installation:
1. install Go
2. go get github.com/gorilla/mux
3. go get github.com/bladewheels/betableBlackJack

### Run:
1. cd $GOPATH/src/github.com/bladewheels/betableBlackJack
2. go run *.go

###### If you have been working with Go longer than I have (2 days!) you may know of a better way to Install and/or Run this code e.g. building, PRs welcome!

#### The actors in this drama include:

##### GAME:
- holds game state
- DEALER
- list of PLAYERs, possibly empty or limited to ~~MAX_PLAYERS~~ 1
- DECK

##### DECK:
- deck of cards, from remote API
- CARDs can be dealt to DEALER or a PLAYER's HAND
- can be shuffled

##### CARD:
- a common playing card
- one 4 suits: Clubs, Spades, Heats, Diamonds
- one 13 denominations: Ace, 2, 3...10, Jack, Queen, King

##### HAND:
- a collection of cards that were dealt from the DECK

##### DEALER:
- deals cards from DECK to self or a PLAYER's HAND in the GAME
- judges whether a HAND is > 21 in TOTAL
- stops self-dealing when 17 TOTAL HAND is reached

##### PLAYER:
- can initiate a game of BlackJack
- can play BlackJack by:
  - asking to be HIT ie add another CARD to their HAND
  - asking to STAND ie allow the DEALER the chance to beat the PLAYER's HAND

##### API, the following URIs are used by PLAYER to interact with the game:
- ~~GET: /api/games returns all GAME IDs.~~ not implemented yet
- ~~GET: /api/games/{gameID} returns the state of a GAME.~~ not implemented yet
- POST: /api/games starts a new game; the PLAYER and the DEALER are dealt their hands at this point, returns the GAME state.
- PATCH: /api/games/{gameID}/hit adds another CARD to the PLAYER's HAND, returns the GAME state; that state may indicate BUST for the PLAYER i.e. if their HAND > 21.
- PATCH: /api/games/{gameID}/stand signals the DEALER to complete play, returns the GAME state; the state includes the outcome of the game.

###### Typical gameplay (e.g. w/curl, telnet, Postman, etc.):

- POST to: /api/games, the response models the state of a new GAME, including the HANDs dealt to the DEALER and PLAYER
 - examine the GAME state and identify the GAME ID; decide whether to HIT or STAND and use the ID in the following calls:
- to HIT, PATCH to: /api/games/{gameID}/hit
 - examine the GAME state and identify your HAND; decide whether to HIT or STAND
 - repeat ad naseum or until BUST or STANDing
- to STAND, PATCH to: /api/games/{gameID}/stand

###### Example use of curl against a localhost:
curl -v -H "Accept: application/json" -H "Content-type: application/json" -X POST http://localhost:8080/api/games

##### Thoughts about the design/implementation:

A GAME is created when the /games endpoint is hit, and it is stored in an in-memory channel. Subsequent calls to /hit or /stand dequeue the GAME, operate upon it and re-queue it.

The Cards and Deck(s) come from an external API (https://deckofcardsapi.com/) who also shuffles for us. I've added a retry mechanism in case pulling those items fails the first time.
