# betableBlackJack
Demo API in Go for a single-user BlackJack game

The nature of HTTP is call and response so this gameplay is most straightforward for a single PLAYER at a TABLE; polling (or eg WebSockets, in a future implementation) may be utilized for multi-PLAYER games but this initial design is INTENDED FOR SINGLE PLAYER USE only, due to time-to-market constraints.

The actors in this drama include:

TABLE:
- holds game state, in serialized JSON
- DEALER
- list of PLAYERs, possibly empty or limited to MAX_PLAYERS
- DECK

DECK:
- deck of cards, from remote API
- CARDs can be dealt to DEALER or PLAYER(s) HANDs
- can be shuffled

CARD:
- a common playing card
- one 4 suits: Clubs, Spades, Heats, Diamonds
- one 13 denominations: Ace, 2, 3...10, Jack, Queen, King

HAND:
- a collection of cards that were dealt from the DECK

DEALER:
- deals cards from DECK to self or PLAYER's HANDs at TABLE
- judges whether a HAND is > 21 in TOTAL
- stops self-dealing when 17 TOTAL HAND is reached

PLAYER:
- can initiate or abandon a game of BlackJack at a TABLE
- can play BlackJack by:
  - asking to be HIT ie add another CARD to their HAND
  - asking to STAND ie allow the DEALER (or other PLAYER(s)) to try to beat the PLAYER's HAND

API, the following URIs are used by PLAYER(s) to interact with the game:
- /api/tables w/GET returns all TABLEs; a PLAYER may join a TABLE if a game is not in progress.
- /api/tables/{tableID} w/GET returns the state of a TABLE.
- /api/tables w/POST returns a new TABLE.
- /api/table/{tableID}/game w/POST starts a new game; all the PLAYERs and the DEALER at the TABLE are dealt their hands at this point, returns the TABLE state. Other PLAYERs at the TABLE would need to poll the /api/table GET endpoint to ascertain the state of the TABLE ie what their HAND looks like.
- /api/table/{tableID}/game/{playerID}/hit w/POST adds another CARD to the PLAYER's HAND, returns the TABLE state; that state may indicate GAME_OVER for the PLAYER ie if their HAND > 21.
- /api/table/{tableID}/game/{playerID}/stand w/POST signals the DEALER that it may now complete play with the PLAYER, returns the TABLE state; if there is only 1 PLAYER at the TABLE then the state includes the outcome of the game, otherwise each PLAYER standing would need to poll the /api/table GET endpoint to ascertain the state of the TABLE ie who won.

Thoughts about the design/implementation:

The ensuing TABLE state could be serialized to JSON by each endpoint, stored in a channel for the next endpoint to consume - not a scalable strategy perhaps but good enough for a quick&dirty demo? Wouldn't work for multi-player. Could put pointer to channel in map, but why?!
Otherwise, throw it into an in-memory map, keyed by TABLE GUID, or a persistent data store eg Mongo? for ease of retrieval; might need access-locking if multi-player? Depends if turn-taking is DEALER-allocated or is free-for-all?
