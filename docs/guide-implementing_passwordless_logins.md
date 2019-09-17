# Passwordless Logins

If passwords are an obstacle in your application, you may prefer to offer a passwordless login. An
example of this process might involve users clicking a link from their email.

AuthN takes care of generating a secure one-use token, while safeguarding against user enumeration.
Your application is responsible for delivering the token to the user with something like an email
link.

Users must have an account before they may log in without a password. You may also be interested in
[passwordless signup](guide-implementing_passwordless_signup.md).

## Configuration

* [`APP_PASSWORDLESS_TOKEN_URL`](config.md#app_passwordless_token_url)
* [`PASSWORDLESS_TOKEN_TTL`](config.md#passwordless_token_ttl)

## Implementation

### Backend

Your application must implement an endpoint (secured by SSL & HTTP Basic Auth) that expects a `POST`
webhook with `account_id` and `token` params. It should use the `account_id` to decide where to
deliver the `token`. When it is complete, it must return a 2xx status code or else AuthN will retry
the notification.

For example, a Rails application might use these params to send an email:

```ruby
class AuthnController < ApplicationController
  def passwordless_login
    @user = User.find_by_account_id(params[:account_id])
    AccountMailer.passwordless_login(@user, params[:token]).deliver_later
    head :ok
  end
end
```

### AuthN

Set [APP_PASSWORDLESS_TOKEN_URL](config.md#app_passwordless_token_url) with the full URL of your new
endpoint. For the example above, it might look like:

```bash
# development
APP_PASSWORDLESS_TOKEN_URL=http://localhost:3000/authn/passwordless_login

# production
APP_PASSWORDLESS_TOKEN_URL=https://user:pass@myapp.io/authn/passwordless_login
```

### Frontend

First, create a place for users to begin the process:

1. Create a form where the user may enter an account name.
2. Submit the account name to AuthN.

  NOTE: AuthN always reports success, even if the account has not registered.

Then, create a page that exchanges the token for a session after users click through an email link:

3. When the page loads, submit the `token` to AuthN
4. If successful, the user will be logged in.

## Related Guides

* [Passwordless Signup](guide-implementing_passwordless_signup.md)
