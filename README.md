# Assignment in GO

## Description:

A simple program that fetches the contents of several URLs and returns the contents in reverse order.
The program is written in GO and can be run in a Docker container.

## Running program:

As an application:
* Make sure you have GO install on a machine (https://golang.org/doc/install)
* In terminal: `go run .`
* Use http://localhost:8080/getUrlContents to get contents

For Docker:
* Docker build: `docker-compose build`
* Docker start: `docker-compose up -d`
* Use http://localhost:8080/getUrlContents to get contents

## Used resources:
* https://go.dev/tour/list
* https://medium.com/@gauravsingharoy/asynchronous-programming-with-go-546b96cd50c1
* https://docs.docker.com/language/golang/build-images/
