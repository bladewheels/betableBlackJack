# betableBlackJack

Demo API in Go for a single-user BlackJack game

The nature of HTTP is call and response so this gameplay is most straightforward for a single PLAYER at a TABLE; polling (or eg WebSockets, in a future implementation) could be utilized for multi-PLAYER games but this initial design is INTENDED FOR SINGLE PLAYER USE only, due to time-to-market constraints. Concurrent single-player games are supported.

The actors in this drama include:

GAME:
- holds game state
- DEALER
- list of PLAYERs, possibly empty or limited to ~~MAX_PLAYERS~~ 1
- DECK

DECK:
- deck of cards, from remote API
- CARDs can be dealt to DEALER or a PLAYER's HAND
- can be shuffled

CARD:
- a common playing card
- one 4 suits: Clubs, Spades, Heats, Diamonds
- one 13 denominations: Ace, 2, 3...10, Jack, Queen, King

HAND:
- a collection of cards that were dealt from the DECK

DEALER:
- deals cards from DECK to self or a PLAYER's HAND in the GAME
- judges whether a HAND is > 21 in TOTAL
- stops self-dealing when 17 TOTAL HAND is reached

PLAYER:
- can initiate or abandon a game of BlackJack
- can play BlackJack by:
  - asking to be HIT ie add another CARD to their HAND
  - asking to STAND ie allow the DEALER the chance to beat the PLAYER's HAND

API, the following URIs are used by PLAYER to interact with the game:
- ~~GET: /api/games returns all GAME IDs.~~ not implemented
- ~~GET: /api/games/{gameID} returns the state of a GAME.~~ not implemented
- POST: /api/games starts a new game; the PLAYER and the DEALER are dealt their hands at this point, returns the GAME state.
- POST: /api/games/{gameID}/hit adds another CARD to the PLAYER's HAND, returns the GAME state; that state may indicate BUST for the PLAYER i.e. if their HAND > 21.
- POST: /api/games/{gameID}/stand signals the DEALER to complete play, returns the GAME state; the state includes the outcome of the game.

Typical gameplay (e.g. w/curl, telnet, Postman, etc.):

- POST to: /api/games, the response models the state of a new GAME, including the HANDs dealt to the DEALER and PLAYER
 - examine the GAME state and identify the GAME ID; decide whether to HIT or STAND and use the ID in the following calls:
- to HIT, POST to: /api/games/{gameID}/hit
 - examine the GAME state and identify your HAND; decide whether to HIT or STAND
 - repeat ad naseum or until BUST or STANDing
- to STAND, POST to: /api/games/{gameID}/stand

Thoughts about the design/implementation:

The ensuing TABLE state could be serialized to JSON by each endpoint, stored in a channel for the next endpoint to consume - not a scalable strategy perhaps but good enough for a quick&dirty demo? Wouldn't work for multi-player. Could put pointer to channel in map, but why?!
Otherwise, throw it into an in-memory map, keyed by TABLE GUID, or a persistent data store eg Mongo? for ease of retrieval; might need access-locking if multi-player? Depends if turn-taking is DEALER-allocated or is free-for-all?
