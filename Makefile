PKGS := $(shell glide nv)
ORG := keratin
PROJECT := authn-server
NAME := $(ORG)/$(PROJECT)
VERSION := 1.0.0-rc3

.PHONY: clean
clean:
	rm -rf vendor
	rm -rf dist
	rm -f authn
	rm -f api/views/*.ego.go

# Code generation (ego)
.PHONY:
generate:
	go generate github.com/$(NAME)/api/views

# Fetch dependencies
vendor:
	glide install
	go install

# Build the project
.PHONY: build
build: generate vendor
	mkdir -p dist
	CGO_ENABLED=1 go build -o dist/authn

.PHONY: build-builder
build-builder:
	docker build -f Dockerfile.builder -t $(NAME)-builder .

.PHONY: docker
docker: build-builder
	docker run --rm \
		-v $(PWD):/go/src/github.com/$(NAME) \
		-w /go/src/github.com/$(NAME) \
		$(NAME)-builder \
		make clean build
	docker build --tag $(NAME):latest .

# Run the server
.PHONY: server
server: vendor generate
	docker-compose up -d redis
	DATABASE_URL=sqlite3://localhost/dev \
		REDIS_URL=redis://127.0.0.1:8701/11 \
		go run *.go

# Run tests
.PHONY: test
test: vendor generate
	docker-compose up -d redis mysql
	TEST_REDIS_URL=redis://127.0.0.1:8701/12 \
	  TEST_MYSQL_URL=mysql://root@127.0.0.1:8702/authnservertest \
	  go test $(PKGS)

# Run CI tests
.PHONY: ci
test-ci:
	TEST_REDIS_URL=redis://127.0.0.1/1 \
	  TEST_MYSQL_URL=mysql://root@127.0.0.1/test \
	  go test -race $(PKGS)

# Run benchmarks
.PHONY: benchmarks
benchmarks:
	docker-compose up -d redis
	TEST_REDIS_URL=redis://127.0.0.1:8701/12 \
		go test -run=XXX -bench=. \
			github.com/keratin/authn-server/api/meta \
			github.com/keratin/authn-server/api/sessions

# Run migrations
.PHONY: migrate
migrate:
	go run *.go migrate

# Cut a release of the current version.
.PHONY: release
release: test docker
	docker tag $(NAME):latest $(NAME):$(VERSION)
	docker push $(NAME):$(VERSION)
	git tag v$(VERSION)
	git push --tags
	open https://github.com/$(NAME)/releases/tag/v$(VERSION)
