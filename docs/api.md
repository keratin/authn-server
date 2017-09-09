---
title: Server API
---

# Server API

* [Visibility](#visibility)
* [JSON Envelope](#json-envelope)
* Endpoints
  * Accounts
    * [Signup](#signup)
    * [Username Availability](#username-availability)
    * [Lock Account](#lock-account)
    * [Unlock Account](#unlock-account)
    * [Archive Account](#archive-account)
    * [Import Account](#import-account)
  * Sessions
    * [Login](#login)
    * [Refresh Session](#refresh-session)
    * [Logout](#logout)
  * Passwords
    * [Request Password Reset](#request-password-reset)
    * [Change Password](#change-password)
    * [Expire Password](#expire-password)
  * Other
    * [Service Configuration](#service-configuration)
    * [JSON Web Keys](#json-web-keys)
    * [Service Stats](#service-stats)

## Visibility

AuthN exposes both **public** and **private** endpoints.

**Public** endpoints are intended to receive traffic directly from a client, although you may certainly route that traffic through a gateway if you prefer. These endpoints rely on trusted HTTP Referer† headers to prevent CSRF attacks. V1.0 will also include support for a custom `AUTHN-AUDIENCE` header, intended for native clients.

† The Referer header was found to exist for [99.9% of users over HTTPS](http://seclab.stanford.edu/websec/csrf/csrf.pdf) (which you should be using anyway).

**Private** endpoints are intended to receive only traffic from your application's backend. They require HTTP Basic Auth username and password, and should only be accessed over HTTPS (which you should be using anyway).

## JSON Envelope

Successful actions will be indicated with a HTTP 2xx code, and usually accompanied by a JSON response containing a `result` key.

Example:

```json
{
  "result": {
    "id_token": "..."
  }
}
```

Failed actions will be indicated with a 4xx or 5xx HTTP code, and accompanied by a JSON response containing an `errors` key that maps to an array of `{field: , message: }` objects.

Example:

```json
{
  "errors": [
    {"field": "username", "message": "TAKEN"},
    {"field": "password", "message": "INSECURE"}
  ]
}
```

## Endpoints

### Signup

Visibility: Public

`POST /accounts`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | Must be present and unique. |
| `password` | string | Must meet minimum complexity scoring per [zxcvbn](https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/). |

#### Success:

    201 Created

    {
      "result": {
        "id_token": "..."
      }
    }

#### Failure

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "username", "message": "MISSING"},
        {"field": "username", "message": "FORMAT_INVALID"},
        {"field": "username", "message": "TAKEN"},
        {"field": "password", "message": "MISSING"},
        {"field": "password", "message": "INSECURE"}
      ]
    }

The reason for `FORMAT_INVALID` will depend on whether you've configured AuthN to validate usernames
as email addresses.

### Username Availability

Visibility: Public

`GET /accounts/available`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | &nbsp; |

#### Success:

    200 Ok

    {
      "result": true
    }

#### Failure

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "username", "message": "TAKEN"}
      ]
    }

### Lock Account

Visibility: Private

`PATCH|PUT /accounts/:id`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Unlock Account

Visibility: Private

`PATCH|PUT /accounts/:id`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Archive Account

Visibility: Private

`DELETE /accounts/:id`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Import Account

Visibility: Private

`POST /accounts/import`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | Must exist and be unique, but otherwise not validated. |
| `password` | string | May be either an existing BCrypt hash or a plaintext (raw) string. Will not be validated for complexity. |
| `locked` | boolean | Optional. Will import the account as [locked](#lock-account). |

#### Success:

    201 Created

    {
      "result": {
        "id": 123456789
      }
    }

#### Failure:

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "username", "message": "MISSING"},
        {"field": "password", "message": "MISSING"}
      ]
    }

### Login

Visibility: Public

`POST /session`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | &nbsp; |
| `password` | string | &nbsp; |

#### Success:

    201 Created

    {
      "result": {
        "id_token": "..."
      }
    }

#### Failure:

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "credentials", "message": "FAILED"},
        {"field": "credentials", "message": "EXPIRED"},
        {"field": "account", "message": "LOCKED"}
      ]
    }

Note that no information is given to tell the user whether the username was found or the password was incorrect.

When handling the `EXPIRED` error for credentials, instruct the user their password must be reset.

### Refresh Session

Visibility: Public

`GET /session/refresh`

As long as a device remains logged in to the AuthN server, it can hit this endpoint to fetch a fresh JWT session. The [`keratin/authn-js`](https://github.com/keratin/authn-js) library can automate this by pre-emptively refreshing tokens when they reach halflife.

This refresh scheme is necessary so that device sessions may be permanently and effectively revoked.

#### Success:

    201 Created

    {
      "result": {
        "id_token": "..."
      }
    }

#### Failure:

    401 Unauthorized

### Logout

Visibility: Public

`DELETE /session`

When a user signs up or logs in, their device establishes a session with the AuthN service, and within that session is a refresh token. This endpoint will revoke the token and discard the session.

#### Success:

    200 OK

### Request Password Reset

Visibility: Public

`GET /password/reset`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | &nbsp; |

#### Success:

    200 Ok

A webhook will be posted to your application's password reset URI with a request body containing:

| Params | Type | Notes |
| ------ | ---- | ----- |
| `account_id` | integer | Provided for your application to easily find the appropriate user. |
| `token` | JWT | Your application must deliver this to the user, usually by email. This JWT's audience is AuthN, and should be opaque to your application. |

#### Failure:

    200 Ok

Note that success and failure are indistinguishable to the client. Even the webhook is performed in the background, to prevent timing attacks.

### Change Password

Visibility: Public

`POST /password`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `password` | string | Must meet minimum complexity scoring per [zxcvbn](https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/). |
| `token` | JWT | As generated by [Request Password Reset](https://github.com/keratin/authn/wiki/Server-API#request-password-reset). This is optional if the user is currently logged in to AuthN. |

#### Success:

    201 Created

    {
      "result": {
        "id_token": "..."
      }
    }

#### Failure:

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "password", "message": "MISSING"},
        {"field": "password", "message": "INSECURE"},
        {"field": "token", "message": "INVALID_OR_EXPIRED"},
        {"field": "account", "message": "NOT_FOUND"},
        {"field": "account", "message": "LOCKED"}
      ]
    }

Note that `NOT_FOUND` may happen if the account is archived after sending a reset token.

### Expire Password

Visibility: Private

`PATCH|PUT /accounts/:id/expire_password`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

Revokes all of the user's current sessions and flags the account for a required password change on their next login. This will manifest as an expired credentials error on what would normally have been a successful login.

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Service Configuration

Visibility: Public

`GET /configuration`

AuthN is not a fully compliant OpenID Connect service, and therefore does not use the `/.well-known/` directory or publish all OpenID Connect-required fields.

This endpoint is primarily used by backend client libraries to fetch the `jwks_uri` path.

#### Success:

| Params | Type | Notes |
| ------ | ---- | ----- |
| `issuer` | string | Base URL of AuthN service, as configured. |
| `response_types_supported` | array[string] | Always `["id_token"]`. |
| `subject_types_supported` | array[string] | Always `["public"]`. |
| `id_token_signing_alg_values_supported` | array[string] | Always `["RS256"]`. |
| `claims_supported` | array[string] | Always `["iss", "sub", "aud", "exp", "iat", "auth_time"]` |
| `jwks_uri` | string | URL for public key necessary to validate JWTs |

### JSON Web Keys

Visibility: Public

`GET /jwks`

This endpoint is primarily used by backend client libraries to fetch the public key necessary to validate the JWTs this AuthN service issues.

#### Success:

| Params | Type | Notes |
| ------ | ---- | ----- |
| `keys.use` | string | Always `"sig"`. |
| `keys.alg` | string | Always `["RS256"]`. |
| `keys.kty` | string | &nbsp; |
| `keys.kid` | string | &nbsp; |
| `keys.e` | string | &nbsp; |
| `keys.n` | string | &nbsp; |

### Service Stats

Visibility: Public

`GET /stats`

Returns estimated statistics for active users over the last trailing 365 days, 104 weeks, and 60 months. Trims off trailing zero entries in each data set, on the assumption that those days predate your application's launch.

Time periods are labeled in ISO8601 formats:

| Period | Format | Example |
| ------ | ------ | ------- |
| day    | `YYYY-MM-DD` | `2016-01-15` |
| week   | `YYYY-\WWW` | `2016-W02` |
| month  | `YYYY-MM` | `2016-01` |

#### Success:

    200 Ok

    {
      "actives": {
        "daily": {},
        "weekly": {},
        "monthly": {}
      }
    }
