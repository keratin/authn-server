# Changelog

## HEAD

### Improved

* Query optimizations on private admin endpoints.

## 1.5.0

### Added

* Passwordless Logins (aka Magic Links) [#71] - @etruta
* New field: `accounts.last_login_at` [#71] - @etruta
* Windows build

### Changed

* Improved printing for configuration errors

### Fixed

* Uncaught uniqueness violation in `PATCH /account/:id`

## 1.4.1

### Fixed

* connection leak with Postgres adapter [#60] - @shashankmehra

## 1.4.0

### Added

* OAuth authentication via Facebook, GitHub, and Google [#50]
* PostgreSQL support [#47] - @Mohammed90

## 1.3.0

### Added

* Improved (simplified) coordination between multiple AuthN servers when synchronizing keys [#44]

## 1.2.1

### Added

* ability to control location of sqlite3 database [#43] - @akhedrane

### Fixed

* aggressively short wlock timeout on blob store (could result in competing keys)

## 1.2.0

### Added

* Log the actual client IP when deployed behind a proxy [#38]
* Bind a second port with only public routes [#37]

## 1.1.0

### Added

* `GET /accounts/:id` endpoint [#30] - @shashankmehra
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
