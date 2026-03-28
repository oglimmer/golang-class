# Epilog

This chapter gives some ideas what you could improve or change next.

# Idea 1 - Game ending

We haven't added code to support the end of a game yet.

Implement:

* Detect that a game has ended when all players have no booking categories left
* Show the winning player and the total score of all players
* Clean up the database

# Idea 2 - Accounts and resume game

A player wants to create an account, so people can pause and re-join games later on.

We need the following functionality:

* Account creation
* Account login
* Create game within that account
* List all running games for your account
* Re-join a game on your account

You have to create additional GORM model structs to persist the data to the database, you also have to write new REST endpoints, change the game logic and adapt the Vue application.

# Idea 3 - Highscore

Implement a persistent highscore for various date ranges.

We want to have:

* Today's top 10 scores
* This year's top 10 scores
* All time top 10 highscore

You have to:

* Add another GORM model struct to store the highscores
* Add REST endpoints to retrieve the highscores for a certain date range type
* Change the game logic to store a highscore at the end of a game if applicable
* Change the Vue application to show the highscores

# Idea 4 - Network multi-player

This is for sure the hardest improvement: Make the game a true multi-computer game, playable from different browsers over the Internet, so it's not a "couch co-op" game anymore.

Implement:

* "Create game" only creates the game for yourself and puts it into a "waiting for other players to join" state
* A new endpoint lists all games "waiting for other players to join"
* A new endpoint lets you join a game in "waiting for other players to join"
* A new endpoint starts a game (available for the game creator)
* Add a WebSocket endpoint to your Go application which informs a client that a REST call has to be done (consider using `gorilla/websocket` or `nhooyr.io/websocket`)
* Add another game state management "waiting for other player to play the turn" and implement it into the game logic
* When a player is about to make a turn decision, send an event via the WebSocket connection
* Change the Vue application to support game creation, waiting game list, join game
* Change the Vue application to connect to the WebSocket after joining a game and on receive event execute the related REST call

Again this is a difficult change and requires a certain level of coding experience.
