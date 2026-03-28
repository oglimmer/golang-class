# Chapter 01

Base project setup with Go modules

# Goal

You have a running REST API - without functionality but it returns HTTP responses.

# Context and Knowledge

* Go is widely used in cloud infrastructure, microservices and DevOps tooling. Companies like Google, Uber, Dropbox and Docker use Go extensively
* Go applications are typically backend or server applications. While you can build desktop apps with Go, its strengths are in server-side development
* Go has a powerful standard library for HTTP (`net/http`), but we'll use **Gin** - the most popular Go web framework - which makes building REST APIs much more convenient
* What [REST](https://en.wikipedia.org/wiki/REST) is, is hard to grasp, but what people mean when they say "we have a REST API" is:
    * HTTP(s) as a network protocol - you can use a browser, JavaScript inside the browser or curl to use your REST API
    * It doesn't use cookies to maintain a session between subsequent calls, some REST APIs use cookies for authentication
    * Data is often considered a resource and the URL represents that. To create a user you do POST on "/api/users", to retrieve all users you do GET on "/api/users", to load the user oli you do GET on "/api/users/oli"
    * Usually JSON is used as input/output data format

## Step 1 - Create the project

Unlike Java/Spring where you use a web generator, in Go you simply create a directory and initialize a Go module.

### Initialize the Go module

```bash
mkdir kniffel
cd kniffel
go mod init github.com/oglimmer/kniffel
```

Replace `github.com/oglimmer/kniffel` with your own module path. The module path is typically the repository URL where your code will live.

This creates a `go.mod` file - similar to Java's `pom.xml` but much simpler. It defines the module name and the Go version.

### Install Gin

```bash
go get github.com/gin-gonic/gin
```

This downloads the Gin framework and adds it to your `go.mod` file as a dependency. You'll also see a `go.sum` file appear - this locks the exact versions of all dependencies (like a lock file in npm).

### Understanding Go project structure

Go projects are simpler than Java projects. There's no deep folder hierarchy required:

```
kniffel/
├── go.mod          # module definition and dependencies (like pom.xml)
├── go.sum          # dependency lock file
└── main.go         # entry point
```

For larger projects, a common structure is:

```
kniffel/
├── go.mod
├── go.sum
├── main.go
├── handler/        # REST handlers (like @RestController)
├── service/        # business logic (like @Service)
├── model/          # data structures (like DTOs and entities)
└── repository/     # database access
```

## Step 2 - Write the main entry point

Create a file `main.go`:

```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    // gin.Default() creates a router with Logger and Recovery middleware
    // Logger logs all requests to stdout
    // Recovery recovers from panics and returns a 500
    r := gin.Default()

    // start the server on port 8080
    r.Run(":8080")
}
```

That's it. In Go, the `main` function in the `main` package is the entry point - no annotations, no XML, no framework magic.

### Run it

```bash
go run main.go
```

You should see Gin's startup output showing the server listening on port 8080.

To test if the REST API is listening, open a browser at http://localhost:8080 - you should see a 404 page. That's fine, as we haven't added any endpoints yet, but we know the server is running.

### Build a binary

```bash
go build -o kniffel .
./kniffel
```

This creates a single binary file. No JVM, no JRE, no JAR - just one executable file. This is what you'd deploy to production or put into a Docker container.

## Step 3 - Ways to test a REST API

### The browser

Any GET request can be tested with a browser - as we did above.

But you cannot easily test a POST, PUT, PATCH or DELETE request, so browser testing is not really ideal.

### Command line using curl

I am a fan of the terminal, so I like REST API testing via curl:

```bash
curl "http://localhost:8080/" -v
```

This should show a 404 response. With curl we can also test any request method (POST, PUT, etc.) and we have full control over all headers or parameters in general.

Also pay attention to the output, you can see bytes sent to the server as an "http request": they are marked with ">". And you can see the full "http response" from the server, all headers are marked with "<" and the body is just at the end. This is invaluable knowledge for debugging HTTP/REST development.

### In the browser using Swagger / OpenAPI

We will set up the project in a way that itself offers a web page to test itself. We'll add this in chapter 03.

### Any other REST API testing application

There are many applications out there to test REST APIs. Maybe the most common one is https://www.postman.com/ - feel free to check it out.

## Step 4 - Push to git

Initialize a git repository:

```bash
git init
```

Create a `.gitignore` file:

```
# Binary
kniffel

# IDE
.idea/
.vscode/
```

Create a new repo in your GitHub/GitLab account and push your project there.

# What we've learnt

* Go projects use `go.mod` for dependency management - much simpler than Maven/Gradle
* `go run` compiles and runs, `go build` creates a binary
* Gin is a popular HTTP framework for Go that provides routing, middleware, and JSON handling
* Go compiles to a single binary - no runtime needed
* REST is based on HTTP and this can be examined in detail with `curl` on the command line

# Extras if you have time

* Read the `go.mod` file - that's Go's dependency file. Compare it to Maven's `pom.xml`
* To see all the project's dependencies, type:

```bash
go list -m all
```

* We can build a production binary:

```bash
# Build for Linux (useful for Docker)
GOOS=linux GOARCH=amd64 go build -o kniffel .

# Build for your current platform
go build -o kniffel .
```

Go supports cross-compilation out of the box - you can build for any platform from any platform.

* You can run the tests (we haven't written any yet):

```bash
go test ./...
```

The `./...` pattern means "all packages in this module and its subdirectories". More on testing in later chapters.
