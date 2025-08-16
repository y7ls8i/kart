## Requirements

* Go (tested on 1.24.6)
* MongoDB (tested on 8 but should work on 7 and maybe 6)

## Dependencies

I use MongoDB because it is easy to set up and use and fast and can handle a lot of data.

I tried to use as minimal libraries as possible:
* `github.com/gin-gonic/gin`         Thin HTTP web framework and HTTP router
* `go.mongodb.org/mongo-driver/v2`   MongoDB driver, obviously
* `github.com/alecthomas/kingpin/v2` For parsing command line arguments
* `github.com/BurntSushi/toml`       For parsing .toml config file
* `github.com/stretchr/testify`      Test helpers library

## Coupon Validation

Coupon validation should not be done against the couponbase files on the fly because it is very slow.
Instead, we should have scripts that select valid coupons from the couponbase files and store them in a MongoDB 
collection.
When serving requests, we should check if the coupon exists in the valid coupons collection.
If the collection is indexed properly, this should be fast and efficient.

The scripts are:
1. `go run ./cmd/dlcoupons` to download couponbase files from S3.
2. `go run ./cmd/validcoupons` to find valid coupons from a list of coupons in data/coupon/ directory.
3. `go run ./cmd/validcoupons` to import valid coupons into a MongoDB collection.

## Structure

#### adapters

Adapters are the interfaces between the application and the external world.
One notable adapter is DB connection. Because I am using MongoDB, therefore the adapter is called adapter/mongo.
Adapters can have other packages, for example if I want to use Redis, I would create an adapter/redis package.

#### business

Business is where the business logic lives. It is a layer above the adapters, meaning business package can have access
to and from the adapters.

#### server

Server is responsible for setting up the HTTP server and routing.
The server package should not have any business logic. It calls business or adapter package to get the data it needs.
If business logic needs to be applied to the data, for example the order creation, it should be done in the business
package.
But if it is simple data fetching, like listing all products, it can be done directly in the adapter package.

#### cmd

Cmd is the entry point of the application. It is responsible for parsing command line arguments, setting up the
configuration, and starting the server.

## Installation

It only has one dependency: MongoDB.
I have created a docker-compose file to make it easy to set up. Obviously, you need Docker already installed.

`docker-compose up -d`

## Tests

`go test ./...`

`go test -race ./...`

## Run

There is a `config.toml` file in the root directory. You can change it to your needs.

`go run ./cmd/api -c config.toml`

## Possible Improvements

In no particular orders:

* Authorization/Authentication
* Observability: logging, metrics, tracing
* Features: products creation, discounts, price calculation, etc.
* GraphQL API
