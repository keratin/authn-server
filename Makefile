PKGS := $(shell glide nv)

.PHONY: clean
clean:
	rm -rf vendor
	rm -rf dist
	rm -f authn

# Fetch dependencies
vendor:
	glide install
	go install

# Build the project
.PHONY: build
build: vendor
	mkdir -p dist
	CGO_ENABLED=1 go build -o dist/authn

.PHONY: docker
docker:
	docker run --rm \
		-v $(PWD):/go/src/github.com/keratin/authn-server \
		-w /go/src/github.com/keratin/authn-server \
		billyteves/alpine-golang-glide:1.2.0 \
		bash -c 'apk add --no-cache gcc musl-dev; make clean build'
	docker build --tag keratin/authn:latest-go .

# Run the server
.PHONY: server
server:
	go run *.go

# Run tests
.PHONY: test
test:
	docker-compose up -d
	TEST_REDIS_URL=redis://127.0.0.1:8701/12 \
	  TEST_MYSQL_URL=mysql://root@127.0.0.1:8702/authnservertest \
	  go test $(PKGS)

# Run migrations
.PHONY: migrate
migrate:
	go run *.go migrate
