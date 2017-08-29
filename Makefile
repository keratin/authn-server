PKGS := $(shell glide nv)

.PHONY: clean
clean:
	rm -rf vendor
	rm -rf dist
	rm -f authn

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

# Fetch dependencies
vendor:
	glide install
	go install

# Run the server
.PHONY: server
server:
	go run *.go

# Run tests
.PHONY: test
test:
	go test $(PKGS)

# Run migrations
.PHONY: migrate
migrate:
	go run *.go migrate
