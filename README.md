Keratin AuthN is an authentication service that keeps you in control of the experience without forcing you to be an expert in web security.

Read more at [keratin.tech](https://keratin.tech).

[![Build Status](https://travis-ci.org/keratin/authn-server.svg?branch=master)](https://travis-ci.org/keratin/authn-server)[![Go Report](https://goreportcard.com/badge/github.com/keratin/authn-server)](https://goreportcard.com/report/github.com/keratin/authn-server)

## Integration

This repository builds a backend Go service that provides secured endpoints related to accounts and passwords. You must integrate it with your application's frontend(s) and backend(s).

Client libraries are currently available for:

* Backends: [Ruby](https://github.com/keratin/authn-rb)
* Frontends: [JavaScript](https://github.com/keratin/authn-js)

If you are missing a client library, please [submit a request](https://github.com/keratin/authn/issues).

### Example: Signup

You will render a signup form as usual, but rely on a provided client library to register the username and password with AuthN in exchange for an ID Token (aka JWT session) that is submitted to your user signup endpoint instead.

    Your Frontend      AuthN       Your Backend
    ===========================================
                 <---------------- signup form

    Email &
    Password     ----> account signup
                 <---- ID Token

    Name &
    Email &
    ID Token     ----------------> user signup

### Example: Password Reset

You will render a reset form as usual, but then submit the username to AuthN. If that username exists, AuthN will communicate a secret token to your application through a secure back channel. Your application is responsible for delivering the token to the user, probably by email.

Your application must then host a form that embeds the token, requests a new password, and submits to AuthN for processing.

    Your Frontend      AuthN       Your Backend
    ===========================================
    Username     ---->
                       account_id &
                       token ---->
                 <---------------- emailed token

                 <---------------- reset form
    Password &
    Token        ----> update

## Configuration

All configuration is through environment variables. Please see [docs](https://github.com/keratin/authn-server/blob/master/docs/config.md) for details.

## Deployment

AuthN may be deployed according to your needs. Here's what it requires:

* A Redis server for session tokens, ephemeral data, and activity metrics.
* A database (currently supports MySQL and Sqlite3) ([submit a request](https://github.com/keratin/authn-server/issues))
* Network routing from your application's clients.
* Network routing to/from your application, for secure back-channel API communication.

In broad strokes, you want to:

1. Provision a server (or decide to colocate it on an existing server)
2. Deploy the code
3. Set environment variables to configure the database and other settings
4. Run migrations
5. Send traffic!

### Maximum Security

For maximum security, give AuthN its own dedicated SQL and Redis databases and be sure that all database backups are strongly encrypted at rest. The credentials and accounts data encapsulated by AuthN should not be necessary for data warehousing or business intelligence, so try to minimize their exposure.

## Developing

1. Install [Glide](https://github.com/Masterminds/glide#install).
2. Run `make vendor` to set up the vendor/ directory using Glide
3. Install Docker and docker-compose.
4. Run `make test` to ensure a clean build

To run a dev server:

1. Create a own `.env` file with desired configuration.
2. Run `make migrate server`

To build a compiled server for integration testing:

1. Run `make build`
2. Execute `dist/authn` with appropriate ENV variables

To build a Docker image for integration testing:

1. Run `make docker`
2. Start the `keratin/authn-server:latest` image with appropriate ENV variables

## COPYRIGHT & LICENSE

Copyright (c) 2016 Lance Ivy

Keratin AuthN is distributed under the terms of the GPLv3. See [LICENSE-GPLv3](LICENSE-GPLv3) for details.
