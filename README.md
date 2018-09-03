# Keratin AuthN

[![Keratin Pangolin](https://keratin.tech/pangolin-logo-dark.gif)](https://keratin.tech)
A modern authentication backend service. ([https://keratin.tech](https://keratin.tech))

[![Gitter](https://badges.gitter.im/keratin/authn-server.svg)](https://gitter.im/keratin/authn-server?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)[![Build Status](https://travis-ci.org/keratin/authn-server.svg?branch=master)](https://travis-ci.org/keratin/authn-server)[![Coverage Status](https://coveralls.io/repos/github/keratin/authn-server/badge.svg)](https://coveralls.io/github/keratin/authn-server)[![Go Report](https://goreportcard.com/badge/github.com/keratin/authn-server)](https://goreportcard.com/report/github.com/keratin/authn-server)

## Related

This repository builds a backend Go service that provides secured endpoints related to accounts and passwords. You must integrate it with your application's frontend(s) and backend(s).

Client libraries are currently available for:

* Backends: [Ruby](https://github.com/keratin/authn-rb) • [Go](https://github.com/keratin/authn-go)
* Frontends: [JavaScript](https://github.com/keratin/authn-js)

If you are missing a client library, please [submit a request](https://github.com/keratin/authn-server/issues).

## Implementation

[Documentation](https://github.com/keratin/authn-server/blob/master/docs/README.md)

## Deployment

[Documentation](https://github.com/keratin/authn-server/blob/master/docs/README.md)

## Configuration

All configuration is through ENV variables.

[Documentation](https://github.com/keratin/authn-server/blob/master/docs/config.md)

## Contributing

Welcome! Please familiarize yourself with the [CONTRIBUTING](CONTRIBUTING.md) doc and the [CODE OF CONDUCT](CODE_OF_CONDUCT.md).

Here's how the dev environment works:

1. `go get github.com/keratin/authn-server`
2. Install Docker and docker-compose.
3. Run `make test` to ensure a clean build

To run a dev server:

1. Create a own `.env` file with desired configuration.
2. Run `make migrate`
3. Run `make server`

To build a compiled server for integration testing:

1. Run `make build`
2. Execute `dist/authn` with appropriate ENV variables

To build a Docker image for integration testing:

1. Run `make docker`
2. Start the `keratin/authn-server:latest` image with appropriate ENV variables

## COPYRIGHT & LICENSE

Copyright (c) 2016-2018 Lance Ivy

Keratin AuthN is distributed under the terms of the LGPLv3. See [LICENSE-LGPLv3](LICENSE-LGPLv3) for details.
