# Login

When a user logs in with AuthN, they establish two sessions: one with your app that expires
periodically, and another with AuthN that can be used to refresh the app session. These are called
the access token and refresh token, respectively.

During login, AuthN works to ensure that users may not enumerate users in your system. This means it
will not declare which field was incorrect, but instead fails with a generic credentials error.

## Configuration

* [ACCESS_TOKEN_TTL](config.md#access_token_ttl)
* [REFRESH_TOKEN_TTL](config.md#refresh_token_ttl)

## Implementation

### Frontend

1. Create a form where the user may enter their username and password.
2. Submit the username and password to AuthN.
3. If successful, the user will be logged in and can make authenticated requests to your app.
