# Server API

* [Visibility](#visibility)
* [JSON Envelope](#json-envelope)
* Endpoints
  * Accounts
    * [Signup](#signup)
    * [Get Account](#get-account)
    * [Update](#update)
    * [Username Availability](#username-availability)
    * [Lock Account](#lock-account)
    * [Unlock Account](#unlock-account)
    * [Delete OAuth account by user id](#delete-oauth-account-by-user-id)
    * [Archive Account](#archive-account)
    * [Import Account](#import-account)

  * Sessions
    * [Login](#login)
    * [Refresh Session](#refresh-session)
    * [Logout](#logout)
    * [Request Passwordless Login](#request-passwordless-login)
    * [Submit Passwordless Login](#submit-passwordless-login)
  * Passwords
    * [Request Password Reset](#request-password-reset)
    * [Change Password](#change-password)
    * [Expire Password](#expire-password)
    * [Password Score](#password-score)
  * OAuth
    * [Begin OAuth](#begin-oauth)
    * [OAuth Return URL](#oauth-return)
    * [Get OAuth accounts info](#get-oauth-accounts-info)
    * [Delete OAuth account](#delete-oauth-account)
  * Multi-Factor Authentication (MFA) **BETA**
    * [New](#totp-new)
    * [Confirm](#totp-post)
    * [Delete](#totp-delete)
  * Other
    * [Service Configuration](#service-configuration)
    * [JSON Web Keys](#json-web-keys)
    * [Service Stats](#service-stats)
    * [Health Check]($health-check)

## Visibility

AuthN exposes both **public** and **private** endpoints.

**Public** endpoints are intended to receive traffic directly from a client, although you may certainly route that traffic through a gateway if you prefer. These endpoints rely on trusted Origin headers to prevent CSRF attacks.

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

Errors might also arise when sending unsupported `Content-Type` headers, or improperly formatted JSON/Form content. In this
case the error message will result in a slightly different payload, accompanied by `400` or `415` Http errors:

```json
{
  "error": "invalid character '}' looking for beginning of value"
}
```

## Endpoints

All PUT / PATCH / POST endpoints support either JSON (`application/json`) or Form (`application/x-www-form-urlencoded`)
encoded requests, depending on which `Content-Type` is found in the request headers. For backwards compatibility,
Authn-Server will try parsing a Form request body when no `Content-Type` is set.

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

### Get Account

Visibility: Private

`GET /accounts/:id`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

#### Success:

    200 Ok

    {
      "result": {
        "id": <id>,
        "username": "...",
        "oauth_accounts": [
          {
            "provider": "google"|"apple",
            "provider_account_id": "91293",
            "email": "authn@keratin.com"
          }
        ],
        "last_login_at": "2006-01-02T15:04:05Z07:00",
        "password_changed_at": "2006-01-02T15:04:05Z07:00",
        "locked": false,
        "deleted": false
      }
    }

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Update

Visibility: Private

`PATCH|PUT /accounts/:id`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |
| `username` | string | &nbsp; |

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "username", "message": "FORMAT_INVALID"}
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

`PATCH|PUT /accounts/:id/lock`

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

`PATCH|PUT /accounts/:id/unlock`

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

### Delete OAuth account by user id

Visibility: Private

`DELETE /accounts/:id/oauth/:name`

|  Params  |    Type   |      Notes      |
| -------- | --------- | --------------- |
| `id`     |  integer  | User account Id |
| `name`   |  string   | Provider names  |

#### Success:

    200 Ok

    {}

#### Failure:
    404 Not Found

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

| Params     | Type   | Notes                               |
|------------|--------|-------------------------------------|
| `username` | string | &nbsp;                              |
| `password` | string | &nbsp;                              |
| `otp`      | string | required if MFA is setup on account |

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
        {"field": "account", "message": "LOCKED"},
        {"field": "otp", "message": "MISSING"},
        {"field": "otp", "message": "INVALID_OR_EXPIRED"},
      ]
    }

> NOTE: no information is given to tell the user whether the username was found or the password was incorrect.

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

### Request Passwordless Login

Visibility: Public

`GET /session/token`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | &nbsp; |

> NOTE: this endpoint only exists when [`APP_PASSWORDLESS_TOKEN_URL`](config.md#app_passwordless_token_url) is configured. If you see a `404 Not Found`, this env variable is missing.

#### Success:

    200 Ok

A webhook will be POSTed to your application's passwordless login URL with a request body containing:

| Params | Type | Notes |
| ------ | ---- | ----- |
| `account_id` | integer | Provided for your application to easily find the appropriate user. |
| `token` | JWT | Your application must deliver this to the user, usually by email. This JWT's audience is AuthN, and should be opaque to your application. |

#### Failure:

    200 Ok

> NOTE: success and failure are indistinguishable to the client. Even the webhook is performed in the background, to prevent timing attacks.

### Submit Passwordless Login

Visibility: Public

`POST /session/token`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `token` | JWT | As generated by [Request Passwordless Login](#request-passwordless-login). |

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
        {"field": "token", "message": "INVALID_OR_EXPIRED"},
        {"field": "account", "message": "NOT_FOUND"},
        {"field": "account", "message": "LOCKED"},
        {"field": "otp", "message": "MISSING"},
        {"field": "otp", "message": "INVALID_OR_EXPIRED"},
      ]
    }

> NOTE: `NOT_FOUND` may happen if the account is archived after sending a passwordless login token.

### Request Password Reset

Visibility: Public

`GET /password/reset`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `username` | string | &nbsp; |

> NOTE: this endpoint only exists when [`APP_PASSWORD_RESET_URL`](config.md#app_password_reset_url) is configured. If you see a `404 Not Found`, this env variable is missing.

#### Success:

    200 Ok

A webhook will be POSTed to your application's password reset URL with a request body containing:

| Params | Type | Notes |
| ------ | ---- | ----- |
| `account_id` | integer | Provided for your application to easily find the appropriate user. |
| `token` | JWT | Your application must deliver this to the user, usually by email. This JWT's audience is AuthN, and should be opaque to your application. |

#### Failure:

    200 Ok

> NOTE: success and failure are indistinguishable to the client. Even the webhook is performed in the background, to prevent timing attacks.

### Change Password

Visibility: Public

`POST /password`

Handles password resets (with token) and password changes (with session).

When [`PASSWORD_CHANGE_LOGOUT`](config.md#password_change_logout) is enabled, all existing sessions for the account will be expired before creating a new session on the current device.

| Params            | Type   | Notes                                                                                                                                     |
|-------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `password`        | string | Must meet minimum complexity scoring per [zxcvbn](https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/). |
| `token`           | JWT    | As generated by [Request Password Reset](#request-password-reset). This is optional if the user is currently logged in to AuthN.          |
| `currentPassword` | string | Must exist when changing a password while logged in (not using token)                                                                     |
| `otp`             | string | required if MFA is setup on account                                                                                                       |

> NOTE: `password` must always be accompanied by _either_ `token` _or_ `currentPassword`.

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
        {"field", "credentials", "message": "FAILED"},
        {"field": "token", "message": "INVALID_OR_EXPIRED"},
        {"field": "account", "message": "NOT_FOUND"},
        {"field": "account", "message": "LOCKED"},
        {"field": "password", "message": "MISSING"},
        {"field": "password", "message": "INSECURE"}
      ]
    }

> NOTE: `NOT_FOUND` may happen if the account is archived after sending a reset token.

### Expire Password

Visibility: Private

`PATCH|PUT /accounts/:id/expire_password`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `id` | integer | available from the JWT `sub` claim |

Revokes all of the user's current sessions, removes their TOTP secret and flags the account for a required password change on their next login. This will manifest as an expired credentials error on what would normally have been a successful login.

#### Success:

    200 Ok

#### Failure:

    404 Not Found

    {
      "errors": [
        {"field": "account", "message": "NOT_FOUND"}
      ]
    }

### Password Score

Visibility: Public

`POST /password/score`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `password` | string | password to be checked for zxcvbn score |

Returns the zxcvbn score and required score set for [`PASSWORD_POLICY_SCORE`](config.md#password_policy_score) for a given password

#### Success:

    200 Ok

    {
      "result": {
        "score": 3,
        "requiredScore": 2
      }
    }

#### Failure:

    422 Unprocessable Entity

    {
      "errors": [
        {"field": "password", "message": "MISSING"}
      ]
    }

### OAuth

OAuth endpoints are enabled for a supported provider when that provider's credentials are [configured](config.md#oauth-clients).

#### Begin OAuth

Visibility: Public

`GET /oauth/:providerName`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `providerName` | string | google |
| `redirect_uri` | URL | Return URL after OAuth. Must be in your application's domain. |

Redirect a user to this URL when you want to authenticate them with OAuth, and include a `redirect_uri` where you want them to return when they're done. From here, a user will proceed to the OAuth provider and back to AuthN's [OAuth Return](#oauth-return) endpoint (as configured with the provider).

#### Success:

    303 See Other
    Location: (OAuth provider)

#### Failure:

    303 See Other
    Location: (redirect URI)

#### OAuth Return

Visibility: Public

`GET /oauth/:providerName/return`

| Params | Type | Notes |
| ------ | ---- | ----- |
| `providerName` | string |
* google
* github
* facebook |

This is the return URL that must be registered with a provider when provisioning credentials. From here, a user will proceed to the `redirect_uri` specified at the [Begin OAuth](#begin-oauth) step.

If the OAuth process failed, the redirect will have `status=failed` appended to the URL.

#### Success:

    303 See Other
    Location: (redirect URI)

#### Failure:

    303 See Other
    Location: (redirect URI with status=failed)

#### Get OAuth accounts info

Visibility: Public

`GET /oauth/accounts`

Returns relevant oauth information for the current session.

#### Success:

    200 Ok

    {
      "result": [
        {
          "provider": "google"|"apple",
          "provider_account_id": "91293",
          "email": "authn@keratin.com"
        }
      ]
    }

#### Failure:

    401 Unauthorized

#### Delete OAuth account

Visibility: Public

`DELETE /oauth/:providerName`

|     Params     |  Type  |  Notes |
| -------------- | ------ | ------ |
| `providerName` | string | google |

Delete an OAuth account from the current session. If the session was initiated via the OAuth flow, the endpoint will returns an error requesting user to reset password.

#### Success:

    200 Ok

    {}

#### Failure:

    401 Unauthorized

### Multi-Factor Authentication (MFA)

**NOTE** - AuthN MFA support is currently considered in beta.  The API will not be considered stable until v2.

#### New:
Visibility: Public

`POST /totp/new`

#### Success:

    200 Ok

    {
      "result": {
        "secret": "XXXXXXXXXXXXX",
        "url": "otpauth://xxxxxxxxxxxxxxxxxxxx",
      }
    }

#### Failure:

    401 Unauthorized
    422 Unprocessable Entity

#### Confirm:
Visibility: Public

`POST /totp/confirm`

| Params | Type   | Notes     |
|--------|--------|-----------|
| `otp`  | string | Required. |                                       

#### Success:

    200 Ok

#### Failure:

    401 Unauthorized
    422 Unprocessable Entity

#### Delete: 
Visibility: Public

`DELETE /totp`

#### Success:

    200 Ok

#### Failure:

    401 Unauthorized
    422 Unprocessable Entity

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

Visibility: Private

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

### Server Stats

Visibility: Private

`GET /metrics`

Returns server stats (memory usage) and traffic stats (counts, timings) that may be used to monitor and alert on server health. The data is formatted for consumption by Prometheus-compatible collectors.

#### Success:

    200 Ok

    # HELP go_goroutines Number of goroutines that currently exist.
    # TYPE go_goroutines gauge
    go_goroutines 10
    [...]
    # HELP http_requests_total How many HTTP requests processed, partitioned by name and status code
    # TYPE http_requests_total counter
    http_requests_total{code="200",name="GET /health"} 97

### Health Check

Visibility: Public

`GET /health`

Returns a JSON hash with key health indicators. This is the intended endpoint for determining if the system is up.

#### Success:

    200 Ok

    {
      "http": true,
      "db": true,
      "redis": false
    }
