# Chapter 04

Web and CLI Clients

# Goal

Developing and testing a REST API via curl, Postman and a (very simple) Vue web application.

# Context and Knowledge

* Having an efficient development workflow is key for successful software development, thus being able to test and iterate on your REST APIs needs to be easy
* Always try to automate your workflows, curl and Postman enable that
* Professional development uses Test Driven Development (TDD) and we will look at automated tests in chapter 05

# Follow up from Chapter 03

After completing the `GameService` and integrating it into `GameHandler` we have basically completed the Kniffel game REST API.

This is how a `GameService` could look like:

```go
type GameService struct {
    // fake database backend
    mu    sync.RWMutex
    games map[string]*model.KniffelGame
}

func NewGameService() *GameService {
    return &GameService{
        games: make(map[string]*model.KniffelGame),
    }
}

func (s *GameService) CreateGame(playerNames []string) *model.KniffelGame {
    players := make([]model.KniffelPlayer, len(playerNames))
    for i, name := range playerNames {
        players[i] = model.KniffelPlayer{Name: name}
    }
    game := model.NewKniffelGame(players)
    s.mu.Lock()
    s.games[game.GameID] = game
    s.mu.Unlock()
    return game
}

func (s *GameService) GetGameInfo(gameID string) *model.KniffelGame {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.games[gameID]
}

func (s *GameService) Roll(game *model.KniffelGame, diceToKeep []int) {
    game.ReRollDice(diceToKeep)
}

func (s *GameService) BookRoll(game *model.KniffelGame, bookingType string) {
    game.BookDiceRoll(bookingType)
}
```

This is how a `GameHandler` could look like:

```go
// GameServiceInterface defines what the handler needs from the service layer
type GameServiceInterface interface {
    CreateGame(playerNames []string) *model.KniffelGame
    GetGameInfo(gameID string) *model.KniffelGame
    Roll(game *model.KniffelGame, diceToKeep []int)
    BookRoll(game *model.KniffelGame, bookingType string)
}

type GameHandler struct {
    gameService GameServiceInterface
}

func NewGameHandler(gameService GameServiceInterface) *GameHandler {
    return &GameHandler{gameService: gameService}
}

func (h *GameHandler) RegisterRoutes(r *gin.RouterGroup) {
    r.POST("/", h.CreateGame)
    r.GET("/:gameID", h.GetGameInfo)
    r.POST("/:gameID/roll", h.Roll)
    r.POST("/:gameID/book", h.Book)
}

func (h *GameHandler) CreateGame(c *gin.Context) {
    var req model.CreateGameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    game := h.gameService.CreateGame(req.PlayerNames)
    c.JSON(http.StatusOK, mapGameResponse(game))
}

func (h *GameHandler) GetGameInfo(c *gin.Context) {
    gameID := c.Param("gameID")
    game := h.gameService.GetGameInfo(gameID)
    c.JSON(http.StatusOK, mapGameResponse(game))
}

func (h *GameHandler) Roll(c *gin.Context) {
    gameID := c.Param("gameID")
    var req model.DiceRollRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    game := h.gameService.GetGameInfo(gameID)
    h.gameService.Roll(game, req.DiceToKeep)
    c.JSON(http.StatusOK, mapGameResponse(game))
}

func (h *GameHandler) Book(c *gin.Context) {
    gameID := c.Param("gameID")
    var req model.BookRollRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    game := h.gameService.GetGameInfo(gameID)
    h.gameService.BookRoll(game, req.BookingType)
    c.JSON(http.StatusOK, mapGameResponse(game))
}
```

## Why do we separate between `GameHandler` and `GameService`?

One could argue that the `GameService` is not needed and you could easily put all the code directly into `GameHandler`. While it might not be obvious in this small example, in larger applications it is very important to separate the REST / web / HTTP aspects of the code from the business / game logic.

The two main reasons you need to have the code separated into Handler and Service are:

* DB transactions
* Re-usable business / game logic

## What you need to keep in mind

* `Handler` structs
    * Handle mapping REST DTOs into logic structs and vice versa
    * Handle HTTP status codes, map errors to status codes - we will see how this works in chapter 05
* `Service` structs
    * Connect the REST layer to business/game logic and the database layer - in our case `games map[string]*KniffelGame` acts as an in-memory database
    * Do not know anything about REST / HTTP, do not work with HTTP status codes or REST DTOs

We should now have a working REST API for Kniffel. Let's test it.

# Step 1 - Testing in Postman

Postman is a REST API testing tool. You can download it here: https://www.postman.com/downloads/

Unfortunately you need an account to use Postman properly and on top of that, I find Postman the wrong tool. It doesn't integrate into CI/CD pipelines, so I automate things with curl. But many people use Postman, so I want to show it here as well.

After starting Postman and logging into your account you can import the OpenAPI / Swagger definition into Postman. Under "Collections" click the "Import" button. Use your OpenAPI definition URL `http://localhost:8080/swagger/doc.json`. You can import it either way.

There are many tutorials and videos on the internet about Postman. You can play the game via Postman, if you like.

# Step 2 - Playing the game via command line with curl+jq

We can write a "real" client for the command line to play Kniffel with `curl` and `jq`.

## Prerequisites for Linux and macOS

On Linux and macOS install the packages `curl` and `jq` with your package manager.

## Prerequisites for Windows

On Windows curl is installed by default and you can install jq via `winget install jqlang.jq`. Don't use "PowerShell", you have to use "CMD" as `curl` has a completely different syntax in PowerShell.

## Creating a game

Execute this on the terminal:

```bash
curl "http://localhost:8080/api/v1/game/" -d '{"playerNames": ["oli","mike"]}' -H "Content-Type: application/json"
```

This should return some JSON, most importantly you need to find the Game ID. This will also show you the dice this player has initially rolled.

A result could look like this:

```JSON
{"gameId":"a1b2c3d4e5f6","playerData":[{"name":"oli","score":0},{"name":"mike","score":0}],"currentPlayerName":"oli","state":"ROLL","usedBookingTypes":[],"availableBookingTypes":["ONES","TWOS","THREES","FOURS","FIVES","SIXES","THREE_OF_A_KIND","FOUR_OF_A_KIND","FULL_HOUSE","SMALL_STRAIGHT","LARGE_STRAIGHT","KNIFFEL","CHANCE"],"diceRolls":[1,1,2,5,5],"rollRound":1}
```

Now the user has to call the re-roll endpoint twice, also passing the dice to keep as a parameter.

## Calling /roll

```bash
# replace a1b2c3d4e5f6 with your game id
# replace 5,5 with the dice values you want to keep - thus not re-roll
curl "http://localhost:8080/api/v1/game/a1b2c3d4e5f6/roll" -X POST -d '{"diceToKeep": [5,5]}' -H "Content-Type: application/json"
```

Now you should see JSON again. The "dice rolls" should have changed, except for the dice you wanted to keep. Call this endpoint a second time.

With the last re-roll you will see that the field state has changed to "BOOK":

```JSON
{"gameId":"a1b2c3d4e5f6","playerData":[{"name":"oli","score":0},{"name":"mike","score":0}],"currentPlayerName":"oli","state":"BOOK","usedBookingTypes":[],"availableBookingTypes":["ONES","TWOS","THREES","FOURS","FIVES","SIXES","THREE_OF_A_KIND","FOUR_OF_A_KIND","FULL_HOUSE","SMALL_STRAIGHT","LARGE_STRAIGHT","KNIFFEL","CHANCE"],"diceRolls":[1,3,5,5,5],"rollRound":3}
```

This means that the next call needs to use the .../book endpoint.

## Calling /book

To call the book endpoint we use another curl command like this:

```bash
# make sure you use a bookingType to score the max points, as I rolled 3x a 5, I go for the category "FIVES"
curl "http://localhost:8080/api/v1/game/a1b2c3d4e5f6/book" -X POST -d '{"bookingType": "FIVES"}' -H "Content-Type: application/json"
```

The result is JSON again and you will see that the currentPlayer has changed. So we need to re-roll dice again.

## Making an easy to use script on Linux and macOS

We can put everything together and make it playable.

Save this as "play.sh", give it execution permissions `chmod +x play.sh` and run it:

```bash
#!/bin/sh

GAME_CREATE=$(curl "http://localhost:8080/api/v1/game/" -d '{"playerNames": ["oli","mike"]}' -H "Content-Type: application/json" -s | jq)

echo "$GAME_CREATE"

GAME_ID=$(echo "$GAME_CREATE" | jq -r '.gameId')

RUNNING=true

while "$RUNNING" = "true"; do
    ROLL_ROUND=0
    while [ $ROLL_ROUND -ne 3 ]; do
      curl "http://localhost:8080/api/v1/game/$GAME_ID" -s | jq

      echo "Enter dice to keep: (comma separated) "
      read -r data

      ROLL_RESPONSE=$(curl "http://localhost:8080/api/v1/game/$GAME_ID/roll" -X POST -d "{\"diceToKeep\": [$data]}" -H "Content-Type: application/json" -s)
      ROLL_ROUND=$(echo "$ROLL_RESPONSE" | jq -r '.rollRound')
    done

    curl "http://localhost:8080/api/v1/game/$GAME_ID" -s | jq

    echo "Enter booking type: "
    read -r data

    curl "http://localhost:8080/api/v1/game/$GAME_ID/book" -X POST -d "{\"bookingType\": \"$data\"}" -H "Content-Type: application/json" -s | jq

done
```

Playing Kniffel via your REST API on the terminal ;)

## Making an easy to use script on Windows

Create a file `play.bat` and put this content into it:

```bat
@echo off
setlocal enabledelayedexpansion

for /f %%i in ('curl http://localhost:8080/api/v1/game/ -d "{\"playerNames\": [\"oli\",\"mike\"]}" -H "Content-Type: application/json"') do set GAME_CREATE=%%i

echo %GAME_CREATE%

for /f %%i in ('"echo %GAME_CREATE%" ^| jq -r .gameId') do set "GAME_ID=%%i"

set RUNNING=true

:while_loop
if "!RUNNING!"=="true" (
    set ROLL_ROUND=0
    :roll_round
    if !ROLL_ROUND! neq 3 (
        curl http://localhost:8080/api/v1/game/%GAME_ID% -s | jq

        set /p DATA="Enter dice to keep: (comma separated)"

        for /f %%i in ('curl http://localhost:8080/api/v1/game/%GAME_ID%/roll -X POST -d "{\"diceToKeep\": [!DATA!]}" -H "Content-Type: application/json" -s ^| jq -r ".rollRound"') do set ROLL_ROUND=%%i

        goto :roll_round
    )

    curl http://localhost:8080/api/v1/game/%GAME_ID% -s | jq

    set /p DATA="Enter booking type:"

    curl http://localhost:8080/api/v1/game/%GAME_ID%/book -X POST -d "{\"bookingType\": \"!DATA!\"}" -H "Content-Type: application/json" -s | jq

    goto :while_loop
)
```

## Why curl + jq matter?

While writing a terminal client looks a bit superfluous or without a real world use-case, it is the testing and automation capabilities of curl, jq and other command line tools which makes them invaluable for software development.

# Step 3 - Writing a real HTML/CSS/JavaScript client

Let's do a crash course on Vue.

This requires that you have Node.js installed. See [their webpage](https://nodejs.org/en/download/package-manager) how to download and install Node.js.

We start by creating a "vue" project via:

```bash
npm create vue@latest
```

You will be asked a couple of questions, feel free to use my answers:

```shell
Vue.js - The Progressive JavaScript Framework

✔ Project name: … kniffel-client
✔ Add TypeScript? … Yes
✔ Add JSX Support? … No
✔ Add Vue Router for Single Page Application development? … No
✔ Add Pinia for state management? … No
✔ Add Vitest for Unit Testing? … No
✔ Add an End-to-End Testing Solution? › No
✔ Add ESLint for code quality? … Yes
✔ Add Prettier for code formatting? … Yes
```

You should be able to:

```bash
cd kniffel-client # step into the newly created directory
npm install # this installs the dependencies for this project
npm run dev # this starts vue for development with a test webserver running on 5173
```

Access the project at http://localhost:5173 - you should see a template project using Vue.

## Creating the HTTP helper classes

As we have a Swagger/OpenAPI specification for our REST API, we don't have to write the TypeScript code for that manually. We can generate this TypeScript code based on the Swagger/OpenAPI specification.

We need to install a generator tool called "openapi-typescript" and its runtime "openapi-fetch":

```bash
npm i openapi-fetch
npm i --save-dev openapi-typescript
```

Now we can generate the code.

*IMPORTANT*:
First we have to make sure that our REST API - the Go application - is started.

Now we download the OpenAPI specification as JSON and save it to `api.json`. Then we run the code generator to create TypeScript code based on the api.json file.

```bash
curl "http://localhost:8080/swagger/doc.json" > api.json
npx openapi-typescript api.json -o ./src/api/v1.d.ts
```

If you look into `./src/api/v1.d.ts` you will find all types and all code to call the HTTP endpoints.

## Create game

We are not looking at all details of Vue / npm / Node.js programming, as this needs its own tutorial. Let's focus on the changes we need to make this project a Kniffel REST API client.

Open the file `src/App.vue`, remove its content.

Let's start with "create game" functionality:

```vue
<script setup lang="ts">
// import a function from the vue framework and the openapi-fetch runtime plus generated classes
import { ref } from 'vue'
import type { Ref } from 'vue'
import createClient from 'openapi-fetch';
import type { components, paths } from '@/api/v1';

// this creates an http client for our REST api.
const client = createClient<paths>({ baseUrl: "http://localhost:8080" });

// define the interface to hold the player name
interface PlayerInformation {
    index: number;
    name: string;
}
// ref() defines a reactive variable
const names = ref<PlayerInformation[]>([
    {index: 0, name: ''},
    {index: 1, name: ''}
]);
// a function to call the REST API using the generated code
async function createGame() {
    const { data } = await client.POST("/api/v1/game/", {
        body: {
            playerNames: names.value.map(n => n.name)
        }
    });
    alert(data);
}
</script>

<template>
    <h1>Player Names</h1>
    <form>
        <ul>
            <li v-for="ply in names" :key="ply.index">
                Player {{ ply.index+1 }}'s name: <input type="text" v-model="ply.name" />
            </li>
        </ul>
    </form>
    <button @click="names.push({index: names.length, name: ''})">Add Name</button> &nbsp;
    <button @click="createGame">Create Game</button>
</template>
```

You should be able to create a game at http://localhost:5173. After pressing the "Create Game" button, you should see response JSON.

### Show the game information

We need to define more data interfaces to use the response of the create game REST call and show it on the screen.

```vue
<script setup lang="ts">
// keep everything else in <script>

// this defines a variable to hold the loaded GameData using a type from the generated code
const gameData : Ref<components["schemas"]["GameResponse"] | undefined> = ref();

// look for this method and change it
async function createGame() {
    const { data } = await client.POST("/api/v1/game/", {
        body: {
            playerNames: names.value.map(n => n.name)
        }
    });
    gameData.value = data;
}
</script>
<!-- replace everything from here on -->
<template>
    <div v-if="!gameData?.gameId">
        <h1>Player Names</h1>
        <form>
            <ul>
                <li v-for="ply in names" :key="ply.index">
                    Player {{ ply.index+1 }}'s name: <input type="text" v-model="ply.name" />
                </li>
            </ul>
        </form>
        <button @click="names.push({index: names.length, name: ''})">Add Name</button> &nbsp;
        <button @click="createGame">Create Game</button>
    </div>
    <div v-if="gameData?.gameId">
        <h1>Game Scores</h1>
        <ul>
            <li v-for="ply in gameData.playerData" :key="ply.name">
                Player {{ ply.name }} - Score: {{ ply.score }}
            </li>
        </ul>
        <h1 class="mt-20">
            Current player: {{ gameData.currentPlayerName }}
        </h1>
    </div>
    <div v-if="gameData?.state === 'ROLL'">
        <h3>Roll round: {{ gameData.rollRound }}</h3>
        <h3 style="margin-top: 30px;">Select the dice to keep:</h3>
          <ul>
              <li v-for="(die, idx) in gameData.diceRolls" :key="idx">
                  {{ die }}
              </li>
          </ul>
    </div>
</template>
```

Now we can see the game information after the game's creation.

## Re-roll the dice

Let's add checkboxes for each die (to keep it) and a button to do the re-roll.

```vue
<script setup lang="ts">
// .. keep everything here

// this will store the checkbox selection
const rerollSelection = ref([false, false, false, false, false]);

// add at the end of script
async function reroll() {
    const diceToKeep : number[] = [];
    if (gameData.value) {
        for (let i = 0; i < gameData.value.diceRolls.length; i++) {
            if (rerollSelection.value[i]) {
                diceToKeep.push(gameData.value.diceRolls[i]);
            }
        }
        const { data } = await client.POST("/api/v1/game/{gameId}/roll", {
            params: {
                path: {
                    gameId: gameData.value.gameId
                }
            },
            body: {
                diceToKeep
            }
        });
        gameData.value = data;
        if (gameData.value) {
            for (let i = 0; i < gameData.value.diceRolls.length; i++) {
                const idxToKeep = diceToKeep.indexOf(gameData.value.diceRolls[i]);
                if (idxToKeep === -1) {
                    rerollSelection.value[i] = false;
                } else {
                    rerollSelection.value[i] = true;
                    diceToKeep.splice(idxToKeep, 1);
                }
            }
        }
    }
}
</script>
<!-- replace everything from here on -->
<template>
    <div v-if="!gameData?.gameId">
        <h1>Player Names</h1>
        <form>
            <ul>
                <li v-for="ply in names" :key="ply.index">
                    Player {{ ply.index+1 }}'s name: <input type="text" v-model="ply.name" />
                </li>
            </ul>
        </form>
        <button @click="names.push({index: names.length, name: ''})">Add Name</button> &nbsp;
        <button @click="createGame">Create Game</button>
    </div>
    <div v-if="gameData?.gameId">
        <h1>Game Scores</h1>
        <ul>
            <li v-for="ply in gameData.playerData" :key="ply.name">
                Player {{ ply.name }} - Score: {{ ply.score }}
            </li>
        </ul>
        <h1 class="mt-20">
            Current player: {{ gameData.currentPlayerName }}
        </h1>
    </div>
    <div v-if="gameData?.state === 'ROLL'">
        <h3>Roll round: {{ gameData.rollRound }}</h3>
        <div> These types are still available:
            {{ gameData.availableBookingTypes }}
        </div>
        <h3 style="margin-top: 30px;">Select the dice to keep:</h3>
          <ul>
              <li v-for="(die, idx) in gameData.diceRolls" :key="idx">
                  {{ die }} <input type="checkbox" v-model="rerollSelection[idx]" />
              </li>
          </ul>
        <button @click="reroll">Roll</button>
    </div>
</template>
```

## Select booking type

Now we need to add a dropdown box to select the booking type and a button to call the REST API to send it.

You can replace everything with this content.

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { Ref } from 'vue'
import createClient from 'openapi-fetch';
import type { components, paths } from '@/api/v1';

const client = createClient<paths>({ baseUrl: "http://localhost:8080" });

interface PlayerInformation {
    index: number;
    name: string;
}

const names : Ref<PlayerInformation[]> = ref([{index: 0, name: ''}, {index: 1, name: ''}]);
const gameData : Ref<components["schemas"]["GameResponse"]|undefined> = ref();
const rerollSelection = ref([false, false, false, false, false]);
const selectedBookingType : Ref<components["schemas"]["GameResponse"]["usedBookingTypes"]|undefined> = ref();

async function createGame() {
    const { data } = await client.POST("/api/v1/game/", {
        body: {
            playerNames: names.value.map(n => n.name)
        }
    });
    gameData.value = data;
}

async function reroll() {
    const diceToKeep : number[] = [];
    if (gameData.value) {
        for (let i = 0; i < gameData.value.diceRolls.length; i++) {
            if (rerollSelection.value[i]) {
                diceToKeep.push(gameData.value.diceRolls[i]);
            }
        }
        const { data } = await client.POST("/api/v1/game/{gameId}/roll", {
            params: {
                path: {
                    gameId: gameData.value.gameId
                }
            },
            body: {
                diceToKeep
            }
        });
        gameData.value = data;
        selectedBookingType.value = undefined;
        if (gameData.value) {
            for (let i = 0; i < gameData.value.diceRolls.length; i++) {
                const idxToKeep = diceToKeep.indexOf(gameData.value.diceRolls[i]);
                if (idxToKeep === -1) {
                    rerollSelection.value[i] = false;
                } else {
                    rerollSelection.value[i] = true;
                    diceToKeep.splice(idxToKeep, 1);
                }
            }
        }
    }
}

async function book() {
  if (gameData.value) {
    const { data } = await client.POST("/api/v1/game/{gameId}/book", {
      params: {
        path: {
          gameId: gameData.value.gameId
        }
      },
      body: {
        bookingType: selectedBookingType.value
      }
    });
    gameData.value = data;
    rerollSelection.value = [false, false, false, false, false];
  }
}

</script>

<template>
    <div v-if="!gameData?.gameId">
        <h1>Player Names</h1>
        <form>
            <ul>
                <li v-for="ply in names" :key="ply.index">
                    Player {{ ply.index+1 }}'s name: <input type="text" v-model="ply.name" />
                </li>
            </ul>
        </form>
        <button @click="names.push({index: names.length, name: ''})">Add Name</button> &nbsp;
        <button @click="createGame">Create Game</button>
    </div>
    <div v-if="gameData?.gameId">
        <h1>Game Scores</h1>
        <ul>
            <li v-for="ply in gameData?.playerData" :key="ply.name">
                Player {{ ply.name }} - Score: {{ ply.score }}
            </li>
        </ul>
        <h1 class="mt-20">
            Current player: {{ gameData.currentPlayerName }}
        </h1>
    </div>
    <div v-if="gameData?.state === 'ROLL'">
        <h3>Roll round: {{ gameData?.rollRound }}</h3>
        <div> These types are still available:
            {{ gameData?.availableBookingTypes }}
        </div>
        <h3 style="margin-top: 30px;">Select the dice to keep:</h3>
          <ul>
              <li v-for="(die, idx) in gameData.diceRolls" :key="idx">
                  {{ die }} <input type="checkbox" v-model="rerollSelection[idx]" />
              </li>
          </ul>
        <button @click="reroll">Roll</button>
    </div>
    <div v-if="gameData?.state === 'BOOK'">
        <h1>Final dice rolls: {{  gameData.diceRolls }}</h1>
        <div class="mt-20">
          Select the booking type:
        </div>
        <select v-model="selectedBookingType">
            <option v-for="cat in gameData.availableBookingTypes" :key="cat" :value="cat">{{ cat }}</option>
        </select>
        <button @click="book">Book</button>
    </div>
</template>

<style scoped>
button,input {
    margin: 10px;
}
.mt-20 {
    margin-top: 20px;
}
</style>
```

Now you can play Kniffel with your REST API in the browser in a 'couch co-op' style.

# What we've learnt

* Postman as REST API testing tool
* curl and jq can be used to debug, test or automate HTTP calls and thus any REST API
* How to build a very simple Vue application

# Extras if you have time

* A long list of useful curl commands https://curl.se/docs/tutorial.html
* Read more about Vue https://vuejs.org/
