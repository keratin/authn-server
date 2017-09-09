---
title: Make Sessions Timeout from Inactivity
tags:
  - guides
---

# Guide: Make Sessions Timeout from Inactivity

The Keratin AuthN setup involves two sessions: one between the user and your application, and another between the user and AuthN.

The application session comprises a series of access tokens with fixed lifetimes that expire frequently. The appearance of a continuous session is achieved by frequently generating new access tokens, which works as long as the AuthN session remains active.

The AuthN session also expires (and can be revoked), but unlike the access tokens uses an inactivity timer that is reset any time the AuthN session is used to refresh the application session.

This means that in the default setup, a user that has been inactive for a couple of hours may no longer have a current access token but they would remain able to generate one on demand from the still-active AuthN session.

So if you truly want sessions to time out from inactivity, you will need to:

1. Configure the [inactivity timeout on the refresh token](https://github.com/keratin/authn/wiki/Server-Configuration#refresh_token_ttl) to your desired interval, e.g. 10 minutes.
2. Configure the [access token timeout](https://github.com/keratin/authn/wiki/Server-Configuration#access_token_ttl) to match.

If you are using `keratin-authn.cookie.js`, the application session will be automatically refreshed as long as the client device remains active. If the client device stops refreshing (i.e. the user closes the browser tab), you can trust that the refresh token will expire shortly.
