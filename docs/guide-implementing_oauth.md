# OAuth

AuthN integrates the OAuth 2.0 protocol so that you may log users in with any supported OAuth
provider. This integration creates standard AuthN accounts that may additionally have passwords for
normal logins and be locked or archived as usual.

Each AuthN account may have one linked identity for each configured provider.

## Configuration

* [OAuth Clients](config.md#oauth-clients)

## Implementation

You must register an OAuth 2.0 client with your chosen provider(s). This means creating an account
and requesting credentials. While registering, you should be asked to enter a Return URL. The
Return URL is an [AuthN endpoint](api.md#oauth-return) and can be determined by joining the AuthN
server's base URL with the path `/oauth/:providerName/return`. For example:

* `https://authn.example.com/oauth/google/return`

or

* `https://www.example.com/authn/oauth/facebook/return`

### Frontend

1. Create a button or action that redirects users to the [Begin OAuth endpoint](api.md#begin-oauth).
   When redirecting, specify a `redirect_uri` that AuthN can use to send the user back to your app.
2. The route specified by `redirect_uri` can determine if the process succeeded or failed based on
   the possible presence of a `status=failed` query parameter.
3. If the process succeeded, then the user will have session with the AuthN server but will not yet
   have a session (aka access token) with your app, so you must invoke the [Refresh API](api.md#refresh-session).
   The authn-js library offers an `importSession()` function to do this.

Some applications will have user profile information in addition to the AuthN account. In this case,
you must determine if the new session already has a user profile (they just logged in) or needs a
new one (they just signed up) and show them a form to provide the extra details.
