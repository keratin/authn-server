# Changelog

Based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## HEAD

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
