# Changelog

Based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## HEAD

### Added

* Public and private APIs for oauth account visibility and removal - requires migration to record user email on oauth accounts (#253)

## 1.19.0

### Added

* Sign in with Apple oauth support

## 1.18.2

### Fixed

* Remove incomplete "provider specific signing key" feature (#250)

## 1.18.1

### Fixed

* Disallow OAuth linking to an account other than current session's account (#246)
* Replace deprecated gopkg.in/square/go-jose.v2 with github.com/go-jose/go-jose/v3 (#240)

## 1.18.0

### Added

* Beta support for TOTP multi-factor authentication (#220, #230, #231)
* Bump prometheus client to 1.11.1 (#229)
* Bump golang.org/x/crypto to 0.17.0 (#233)

## 1.17.1

### Added

* HMAC notification signatures (#207)
* support inferring default port from configured URL (via #214)
* fix unmarshaling bug in Microsoft OAuth Provider (via #214)

## 1.17.0

### Added

* support for persistent cookies via `REFRESH_TOKEN_EXPLICIT_EXPIRY` env var [#208]

## 1.16.0

### Added

* Usernames may not be passwords [#200]
* ID token contains Session ID claim (`sid`) [#205]

## 1.15.0

### Added

* Added `/jwks` to both public and private routes [#198]

## 1.14.0

### Added

* Added `last_login_at` and `password_changed_at` to Get Account API [#195]

## 1.13.0

### Added

* Support for non-default Redis user [#191]
* Support for TLS connections to Redis with `rediss` [#190]

## 1.12.0

### Added

* Update to go 1.17
* Flexible app domains with wildcard matching [#189]

## 1.11.0

### Added

* Support for Redis Sentinel [#181]

### Fixed

* Improved validation for AUTHN_URL and other ENV url values [#178]

## 1.10.4

### Fixed

* Broken pipe error on Postgres [#174]

## 1.10.3

### Fixed

* Usernames are now case insensitive on Postgres and SQLite. This requires a migration that can fail if the existing database has unintended duplicates! [#170]

## 1.10.2

### Fixed

* CORS configuration allows content-type header

## 1.10.1

### Fixed

* added a timeout to webhook sender

## 1.10.0

### Added

* OAuth through Microsoft [#155]

## 1.9.0

### Added

* endpoint for checking zxcvbn password score [#149]
* option to expire an account's sessions after a password change [#154]

### Fixed

* improvements to constant time comparison in basic auth (thanks @lsmith130)

## 1.8.0

### Added

* Support `Content-Type: application/json` [#143]
* Support for SameSite property on AuthN session cookie [#147]

## 1.7.0

### Added

* OAuth authentication through Discord [#116]

### Fixed

* Email validations no longer allow misplaced periods in the domain

## 1.6.0

### Added

* Log when rejecting a request for a missing or invalid Origin header [#34]
* Accept PUT HTTP calls on every endpoint accepting PATCH [#104]

### Changed

* Same-origin requests are now accepted (for browsers that do not send Origin header for same-origin), by falling back to Referer header to determine the application domain that should be selected in the request's context. The Referer header is only consulted when Origin is not set. Since browsers are only permitted to omit Origin header for same-origin requests this behavior should be robust. [#105]
* Query optimizations on private admin endpoints.
* Pre-compute JWK key on RSA key generation and include within private key wrapper type for use by dependees. [#100]

### Fixed

* panic while evaluating some utf8 password characters
* zxcvbn library we use exhibited some deviation from standard (see: https://github.com/nbutton23/zxcvbn-go/issues/20) so switched to https://github.com/trustelem/zxcvbn [#99]

## 1.5.0

### Added

* Passwordless Logins (aka Magic Links) [#71]
* New field: `accounts.last_login_at` [#71]
* Windows build

### Changed

* Improved printing for configuration errors

### Fixed

* Uncaught uniqueness violation in `PATCH /account/:id`

## 1.4.1

### Fixed

* connection leak with Postgres adapter [#60]

## 1.4.0

### Added

* OAuth authentication via Facebook, GitHub, and Google [#50]
* PostgreSQL support [#47]

## 1.3.0

### Added

* Improved (simplified) coordination between multiple AuthN servers when synchronizing keys [#44]

## 1.2.1

### Added

* ability to control location of sqlite3 database [#43]

### Fixed

* aggressively short wlock timeout on blob store (could result in competing keys)

## 1.2.0

### Added

* Log the actual client IP when deployed behind a proxy [#38]
* Bind a second port with only public routes [#37]

## 1.1.0

### Added

* `GET /accounts/:id` endpoint [#30]
* Airbrake error reporting [#32]
* AuthN version number is now printed on startup

## 1.0.2

### Fixed

* bug with account archival [#29]

## 1.0.1

### Fixed

* Recovery of RSA keys from SQLite3 blob store when restarting AuthN

## 1.0.0

### Added

* AuthN can run entirely from SQLite3 (without Redis)
* LogReporter prints more information to associate an error with a request

### Fixed

* Inverted logic in `GET /accounts/available`
