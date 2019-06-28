# Reset Passwords

When a user forgets their password, the best solution is to send them an email containing a link
that allows them to create a new one. AuthN ensures that the forgotten password process does not
allow attackers to enumerate user accounts while provides a secure reset token to your application
for delivery.

## Configuration

* [PASSWORD_RESET_TOKEN_TTL](config.md#password_reset_token_ttl)
* [APP_PASSWORD_RESET_URL](config.md#app_password_reset_url)
* [PASSWORD_POLICY_SCORE](config.md#password_policy_score)

## Implementation

### Backend

Your application must implement an endpoint (secured by SSL & HTTP Basic Auth) that expects a `POST`
webhook with `account_id` and `token` params. It should use the `account_id` to decide where to
deliver the `token`. When it is complete, it must return a 2xx status code or else AuthN will retry
the notification.

For example, a Rails application might use these params to send an email:

```ruby
class AuthnController < ApplicationController
  def password_reset
    @user = User.find_by_account_id(params[:account_id])
    AccountMailer.password_reset(@user, params[:token]).deliver_later
    head :ok
  end
end
```

### AuthN

Set [APP_PASSWORD_RESET_URL](config.md#app_password_reset_url) with the full URL of your password
reset endpoint. For the example above, it might look like:

```bash
# development
APP_PASSWORD_RESET_URL=http://localhost:3000/authn/password_reset

# production
APP_PASSWORD_RESET_URL=https://user:pass@myapp.io/authn/password_reset
```

### Frontend

First, create a place for users to begin the process:

1. Create a form where the user may enter an account email.
2. Submit the email to AuthN. Note that AuthN always reports success.

Then, create a place for users to continue the process after clicking through your email:

3. Create a form where the user may enter a new password. This form needs the `token` that your app
   sent to the user earlier.
4. Submit the `token` and new password to AuthN.
5. If successful, the user will be logged in.

## Related Guides

* [Displaying a password strength indicator](guide-displaying_a_password_strength_meter.md)
