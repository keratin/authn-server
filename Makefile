include .env
ORG := keratin
PROJECT := authn-server
NAME := $(ORG)/$(PROJECT)
VERSION := 1.19.0
MAIN := main.go

BIN := $(shell pwd)/bin
export GOBIN := $(BIN)
export PATH := $(BIN):$(PATH)

.PHONY: clean
clean:
	rm -rf dist

init: $(BIN)/ego
	go install
	$(BIN)/ego server/views

# Run the server
.PHONY: server
server: init
	docker-compose up -d redis
	DATABASE_URL=sqlite3://localhost/dev \
		REDIS_URL=redis://127.0.0.1:8701/11 \
		go run -ldflags "-X main.VERSION=$(VERSION)" $(MAIN)

# Run tests
.PHONY: test
test: init
	docker-compose up -d redis mysql postgres
	TEST_REDIS_URL=redis://127.0.0.1:8701/12 \
	  TEST_MYSQL_URL=mysql://root@127.0.0.1:8702/authnservertest \
	  TEST_POSTGRES_URL=postgres://$(DB_USERNAME):$(DB_PASSWORD)@127.0.0.1:8703/postgres?sslmode=disable \
	  go test -race ./...

# Run benchmarks
.PHONY: benchmarks
benchmarks:
	docker-compose up -d redis
	TEST_REDIS_URL=redis://127.0.0.1:8701/12 \
		go test -run=XXX -bench=. \
			github.com/keratin/authn-server/server/meta \
			github.com/keratin/authn-server/server/sessions

# Run migrations
.PHONY: migrate
migrate:
	docker-compose up -d redis
	DATABASE_URL=sqlite3://localhost/dev \
		REDIS_URL=redis://127.0.0.1:8701/11 \
		go run -ldflags "-X main.VERSION=$(VERSION)" $(MAIN) migrate

# Cut a release of the current version.
# 1. update CHANGELOG.md and Makefile with the next semantic version
# 2. `make release`
# 3. wait for build and attach artifacts to GitHub release
.PHONY: release
release:
	git push
	git tag v$(VERSION)
	git push --tags
	open https://github.com/$(NAME)/releases/tag/v$(VERSION)

$(BIN)/ego:
	go install github.com/benbjohnson/ego/...

# Build a docker image locally for your current platform.
# TODO: look into buildx support for multi-arch builds
docker-local:
	docker build . -t $(NAME):local -f Dockerfile.local

# Run configured linters
lint:
	@which golangci-lint > /dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin
	golangci-lint run --config .golangci.yml
