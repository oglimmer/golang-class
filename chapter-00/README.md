# Chapter 00

Install Go, an IDE/Editor and Docker

# Goal

You have Go and Docker installed.

# Context and Knowledge

* Go (often called Golang) is a statically typed, compiled language designed at Google
* Go compiles to a single binary - no runtime or virtual machine needed
* Go has a built-in package manager and build tool (`go` command)
* Go is known for its simplicity, fast compilation and great standard library
* Docker makes development in many situations much easier, as it easily provides 3rd party software like webservers or databases

# Step 1

Install Go and either VS Code or GoLand (JetBrains IDE for Go).

* **All platforms**: Download from https://go.dev/dl/ and follow the instructions
* **macOS**: `brew install go`
* **debian/ubuntu**: `sudo apt install golang-go` or download from the official site for the latest version
* **windows**: `winget install GoLang.Go`

For your IDE:
* **VS Code**: Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) - it provides code completion, debugging, linting and more
* **GoLand**: JetBrains' dedicated Go IDE, has everything built-in

# Step 2

You should be able to use the Go toolchain:

```bash
go version
```

You should see something like `go version go1.22.0 darwin/arm64` (version and platform will vary).

Go is both a compiler and a runtime tool. Unlike Java where you have `javac` and `java`, in Go everything is done through the `go` command:

```bash
# compile and run a Go file
go run main.go

# build a binary
go build -o myapp .

# run tests
go test ./...

# manage dependencies
go mod tidy
```

# Step 3

Download Docker Desktop from https://www.docker.com/products/docker-desktop/ and install it.

Docker is very easy to use and doesn't require any additional configuration.

After installing it, here are first steps you might want to try:

```bash
# we start a nginx webserver and map port 80 from inside the container to your host
docker run --rm -p 80:80 nginx
# now try http://localhost/ in your browser
```

or

```bash
# start a container which has curl and jq installed, do a GET http request to math.oglimmer.com to solve 5+7*4 and format the resulting JSON nicely
docker run --rm apteno/alpine-jq /bin/sh -c "curl 'https://math.oglimmer.de/v1/calc?expression=5+7*4' -s | jq"
```

or

```bash
# start a Postgres database with user=postgres, password=postgres
docker run --rm -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres
# download any Postgres client and access your database at localhost with postgres/postgres
```

# What we've learnt

* Go basics and toolchain
* Docker
