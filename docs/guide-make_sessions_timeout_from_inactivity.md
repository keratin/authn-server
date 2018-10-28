# Inactive Session Timeout

The Keratin AuthN setup involves two sessions: one between the user and your application, and another between the user and AuthN.

The application session comprises a series of access tokens with fixed lifetimes that expire frequently. The appearance of a continuous session is achieved by frequently generating new access tokens, which works as long as the AuthN session remains active.

The AuthN session also expires (and can be revoked), but unlike the access tokens uses an inactivity timer that is reset any time the AuthN session is used to refresh the application session.

This means that in the default setup, a user that has been inactive for a couple of hours may no longer have a current access token but they would remain able to generate one on demand from the still-active AuthN session. You must adjust the refresh token's TTL to effectively enforce an inactivity timeout.

## Configuration

* [REFRESH_TOKEN_TTL](config.md#refresh_token_ttl)
* [ACCESS_TOKEN_TTL](config.md#access_token_ttl)

## Implementation

1. Configure the refresh token to your desired timeout, e.g. 10 minutes.
2. Configure the access token timeout to match.

> NOTE:
> If you are using the Keratin AuthN JavaScript library, the application session will be automatically refreshed as long as the client remains open. When the user closes the browser tab, the inactivity timer will begin.
