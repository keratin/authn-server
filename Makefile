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
docker: build
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
