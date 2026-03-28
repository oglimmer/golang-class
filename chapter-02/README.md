# Chapter 02

REST endpoints

# Goal

Basic knowledge of REST endpoints and your REST API exposes some.

# Context and Knowledge

* We want to design our Kniffel REST API as simple as possible, this decision drives some design aspects
    * Our use-case is a "couch co-op" scenario - all players use the same client UI
    * We do NOT separate endpoints for game, player/account, dice
    * Our API isn't concerned about any authentication as all players sit next to each other
* We need the following functionality as REST endpoints (so we need 4 endpoints):
    1. Create a game, specifying the player names
    1. Retrieve the current game information, statistics, options
    1. Decide which dice to re-roll
    1. Decide which category to score the current dice roll
* What's important to know about "REST endpoints"?
    * REST is not a protocol, not a specification, it's a philosophy
    * Still REST uses the HTTP protocol and JSON as input/output data
    * So each **endpoint** is the same as a **URL** - like /api/game or /api/game/37475/roll-dice
    * HTTP (not *REST*) defines several "methods": GET, POST, PATCH, PUT, DELETE - but REST uses this idea to distinguish between create, read, update, delete. It's important to understand, that only the combination of a URL and a method defines an endpoint
    * Once a REST API endpoint is offered to the public, you should not change it, therefore we always want to use a version number in an endpoint - if we want to change something, we increase the version number in the URL
* HTTP methods - conventions for REST:
    * GET - data retrieval of any kind or form
    * POST - creation of a data record / resource or invocation of an action
    * PUT - updating a data record / resource (all fields)
    * PATCH - updating some fields of a data record / resource
    * DELETE - deleting a data record / resource

## Step 1 - Our first endpoint

In Go with Gin, we define endpoints by registering handler functions on the router. There are no annotations like in Java - everything is explicit.

### Add a simple POST endpoint

Let's add an endpoint to our `main.go`:

```go
package main

import (
    "fmt"
    "net/http"
    "runtime"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // group all server endpoints under /api/v1/server
    server := r.Group("/api/v1/server")
    {
        // POST /api/v1/server/goroutine-dump
        server.POST("/goroutine-dump", func(c *gin.Context) {
            // print all goroutine stack traces to stdout
            buf := make([]byte, 1<<16)
            n := runtime.Stack(buf, true)
            fmt.Println(string(buf[:n]))
            // return 200 with no body
            c.Status(http.StatusOK)
        })
    }

    r.Run(":8080")
}
```

Start the application and execute the HTTP request:

```bash
curl "http://localhost:8080/api/v1/server/goroutine-dump" --request POST --verbose
```

As this is a POST request you cannot just put it into the browser address bar (these are always GET requests). You need to use JavaScript. This HTML page uses the fetch API to execute the HTTP request:

```html
<!DOCTYPE html>
<html lang="en">
<body>

<button id="execButton">Execute Request</button>

<script>
document.getElementById('execButton').addEventListener('click', async () => {
    try {
        const response = await fetch("http://localhost:8080/api/v1/server/goroutine-dump", { method: "POST" });
        alert(`Request successful! Return code = ${response.status}. See dev console for more.`);
    } catch (error) {
        console.error('Fetch error:', error);
        alert('Error occurred! See dev console for more.');
    }
});
</script>

</body>
</html>
```

Save this HTML as `test.html` and run:

```bash
docker run --rm -v $PWD:/usr/share/nginx/html -p 80:80 nginx

# on windows
docker run --rm -v %cd%:/usr/share/nginx/html -p 80:80 nginx
```

in the same directory. Access the page with a browser at http://localhost/test.html

To replicate the problem you cannot open test.html directly in your browser, you either have to use Docker or something like xampp.

Does it work? No? Why?

When you check the dev console (what you should always do in those cases), you see the error message "CORS Missing Allow Origin" in the network tab.

## Step 2 - Understanding CORS

* CORS stands for "Cross-Origin Resource Sharing" - but we need to start a step earlier
* Browsers have a "Same-Origin-Policy", which means JavaScript can only do HTTP requests to the **same host and the same port** as your website. In our above example we call from localhost:80 to localhost:8080 - this is forbidden by the Same-Origin-Policy.
* So CORS provides a possibility to overwrite the Same-Origin-Policy
* What's CORS technically? CORS are HTTP response headers sent by the REST API. Our REST API needs to send a header like this:

```
Access-Control-Allow-Origin: *
```

This will tell the browser:

> All websites are allowed to call this REST API.

If you want to be more secure, you can only allow certain websites to call your REST API:

```
Access-Control-Allow-Origin: http://localhost
```

Adding CORS headers in Gin requires a middleware. Install the Gin CORS middleware:

```bash
go get github.com/gin-contrib/cors
```

Now update your `main.go`:

```go
package main

import (
    "fmt"
    "net/http"
    "runtime"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

func main() {
    r := gin.Default()

    // allow all origins (equivalent to @CrossOrigin in Spring)
    r.Use(cors.Default())
    // for more control:
    // r.Use(cors.New(cors.Config{
    //     AllowOrigins: []string{"http://localhost"},
    //     AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    //     AllowHeaders: []string{"Content-Type"},
    // }))

    server := r.Group("/api/v1/server")
    {
        server.POST("/goroutine-dump", func(c *gin.Context) {
            buf := make([]byte, 1<<16)
            n := runtime.Stack(buf, true)
            fmt.Println(string(buf[:n]))
            c.Status(http.StatusOK)
        })
    }

    r.Run(":8080")
}
```

Test the browser page again!

Before we continue, it should be said that CORS has many more aspects, like you need to allow headers and methods individually, also CORS has the concept of preflight requests. But our Kniffel REST API does not require attention to these topics.

## Step 3 - Returning JSON

REST uses JSON as the data format for input and output of data. Let's add a "status" endpoint which returns status information via JSON.

In Go, we define a struct for the response data. Go uses struct tags to control JSON serialization:

```go
// ServerStatus is a simple DTO for the status endpoint
type ServerStatus struct {
    ServerTime    int64  `json:"serverTime"`
    ServerVersion string `json:"serverVersion"`
    ServerName    string `json:"serverName"`
    ServerStatus  string `json:"serverStatus"`
}
```

The `json:"serverTime"` part is called a "struct tag". It tells Go's JSON encoder to use `serverTime` as the JSON field name instead of `ServerTime`. In Go, exported fields (starting with uppercase) are public - the struct tag lets us use camelCase in our JSON output.

Now add the endpoint:

```go
server.GET("/status", func(c *gin.Context) {
    status := ServerStatus{
        ServerTime:    time.Now().UnixMilli(),
        ServerVersion: "1.0.0",
        ServerName:    "Kniffel Server",
        ServerStatus:  "OK",
    }
    // c.JSON automatically serializes the struct to JSON and sets Content-Type
    c.JSON(http.StatusOK, status)
})
```

Don't forget to add `"time"` to your imports.

Test the endpoint with a browser http://localhost:8080/api/v1/server/status or `curl http://localhost:8080/api/v1/server/status --verbose`

## Step 4 - Passing JSON into the REST API

For our Kniffel REST API we want to see how to pass an array of strings - the player names - into a POST call - the create game endpoint.

First, define the request struct:

```go
// CreateGameRequest is the DTO for creating a new game
type CreateGameRequest struct {
    PlayerNames []string `json:"playerNames"`
}
```

Now add the game route group and the create endpoint:

```go
game := r.Group("/api/v1/game")
{
    game.POST("/", func(c *gin.Context) {
        var req CreateGameRequest
        // ShouldBindJSON reads the request body and deserializes JSON into the struct
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        // for now just print it
        fmt.Printf("createGameRequest = %+v\n", req)
        c.Status(http.StatusOK)
    })
}
```

Don't forget to add `"fmt"` to your imports.

You can test this via terminal:

```bash
curl http://localhost:8080/api/v1/game/ --request POST --data-ascii '{"playerNames": ["oli", "ilo"]}' --header "Content-Type: application/json" --verbose
```

On Windows you might need to escape differently:

```bash
curl http://localhost:8080/api/v1/game/ --request POST --data-ascii "{\"playerNames\": [\"oli\", \"ilo\"]}" --header "Content-Type: application/json" --verbose
```

As we send a request body (`--data-ascii` defines the request body), we need to let the server know what data format is used in the body, we do this via the header `Content-Type: application/json`.

## Step 5 - Extend the HTML page to call the "create game" endpoint

Extend the HTML page from above (test.html) to call our new endpoint. You have to make 3 changes:

* Change the URL, the HTTP method should be good
* Provide the header for the content-type = application/json
* Pass the request body (either as a fixed string or get it dynamically from a text field)

You can find some help at https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch or ask ChatGPT.

## Step 6 - Path variables

We have to look at one last building block for our REST endpoints: path variables.

We already saw the endpoint to create a game: POST to /api/v1/game/

To retrieve the data for an existing game the URL should look like this: `/api/v1/game/<ID>` and the method should be a GET request.

```go
game.GET("/:gameID", func(c *gin.Context) {
    // c.Param extracts the path variable from the URL
    gameID := c.Param("gameID")
    fmt.Printf("gameID = %s\n", gameID)
    c.Status(http.StatusOK)
})
```

In Gin, path variables are defined with `:variableName` in the route pattern (similar to `{variableName}` in Spring).

Call this with curl (like `curl http://localhost:8080/api/v1/game/37646`) or a web browser.

## Step 7 - Designing our 4 Kniffel API endpoints

With the knowledge we have learnt from this chapter you should be able to complete or create all 4 endpoints:

1. A POST endpoint for "create game" with a request body containing a string slice for player names. This needs to return at least the game ID, but maybe it can also return the same data as our GET endpoint
1. A GET endpoint to retrieve an existing game with a path variable containing the ID of a game. This should return a struct containing all relevant game information, like player names, their score, the thrown dice and the options to proceed and whatever you think the client needs to know
1. A POST endpoint to re-roll the dice the player hasn't liked. This needs to have a path variable for the game ID and a request body for the dice which should be kept (maybe `[]int`). This should also return the same information as the GET endpoint
1. A POST endpoint to "book" the current dice roll into one of the Kniffel categories. Again we need a path variable for the game ID and a request body for the "booking type" (`string`). This should also return the same information as the GET endpoint

Come up with suggestions for your REST endpoint methods. **It doesn't matter if the result is not perfect - the goal is to develop a proposal.**

# What we've learnt

* In Go/Gin, routes are registered explicitly on the router - no annotations
* Gin concepts
    * `r.Group()` - group routes under a common URL prefix
    * `r.GET()`, `r.POST()` - register handlers for HTTP methods
    * `c.ShouldBindJSON()` - deserialize JSON request body
    * `c.JSON()` - serialize and return JSON response
    * `c.Param()` - extract path variables
    * `cors.Default()` - CORS middleware
* Go struct tags like `json:"fieldName"` control JSON serialization
* What Same-Origin-Policy or CORS is and how to use it
* How to return JSON (response body is JSON) and how to pass JSON into the REST API (request body is JSON)
* How to use path variables - a variable inside a URL

# Extras if you have time

* Read more about browser security https://en.wikipedia.org/wiki/Same-origin_policy and https://en.wikipedia.org/wiki/Cross-origin_resource_sharing
* Read more about Gin's documentation https://gin-gonic.com/docs/
* Explore Go's built-in `net/http` package to understand what Gin provides on top of it
