# Contributing

> Have you found a security issue? Please submit through https://hackerone.com/keratin so that
> a new release may be coordinated with the vulnerability disclosure.

Thanks for taking the time to contribute! Every bit helps.

Before you get too deep in code, please file an issue to discuss the change. Here's a list of
considerations for deciding whether AuthN should be changed:

* Is it a bug? Bugs need fixing. Pull requests, please!
* Is the change backwards compatible? Can it be?
* What effect will the change have on performance?
* Is it within AuthN's scope?

## Scope

### Yes

AuthN's scope of responsibility is currently:

* passwords: verifying, storing, changing
* sessions: creating, securing, revoking

Features such as account archiving, un/locking, and password expiration are all included in the
above scope primarily because they provide the host application with the moderation controls needed
to implement business logic rules that affect the login process. Features that are not in scope (see
below) can usually still be implemented with these atomic moderation actions.

### No

AuthN's scope does not include:

* Delivering notifications to the end-user. AuthN only delivers webhooks.
* Identifying multiple accounts. This responsibility deserves its own service/product.
* Rendering user-facing HTML

## Orientation

### Design

AuthN uses a services pattern to implement the business logic of RESTful route handlers.

* `app/`
  * Configuration routines read from ENV variables to prepare the server.
  * Services encode AuthN's business logic and validations. Any meaningful work in a
    web request should be performed by a service object.
  * Data Access Objects (DAO) implement an interface for pluggable persistence backends
* `server/`
  * The REST implementation of AuthN's HTTP server
  * Routes are responsible for invoking the correct handler
  * Handlers are responsible for translating HTTP requests into service commands, and translating
    service results back into HTTP responses.

### Dependency Injection

* The `main` package resolves an App struct with all configuration and DAOs.
* Handlers are bound to an App struct and may access any config or DAO they need.
* Services get their configuration and DAOs from handlers.

### Error Handling

Errors are passed up from services to handlers. If an error is not recoverable, the handler should
panic (fatal) or report (warning). Panics are handled in middleware that reports them to an outside
service while generating a HTTP 500 error.

## Tests

Yes please!

Any service, API handler, lib package, or data interface method must have tests.

Tests should:

* be colocated with the tested file using a `_test.go` suffix
* prefer sub-tests over table tests (solve redundancy with test helpers)
* use `require.NoError` during test case setup, and `assert.NoError` when testing outcomes
