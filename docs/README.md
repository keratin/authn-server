# Introduction

AuthN is an accounts microservice. It removes passwords and authentication security from your app.

An AuthN account is a login identity: username/password, plus optional connected OAuth identities.

## Why?

* **Stability:** Your application changes every day. Authentication does not.
* **Stack:** Your application is written in a language that encourages moving quickly and fixing
  things later.
* **Security:** Your application's security perimeter depends on the code your team writes and the
  code your team finds.
* **Architecture:** Your application may be small now, but when it grows up you want to be ready.

## Users and Accounts

AuthN manages accounts. When an account logs in, your application receives a token containing the
account's ID. You can save that account ID into your users table along with any other bits of data
that you need like names, time zones, and newsletter preferences.

Integrating AuthN tokens into a typical Ruby on Rails controller with the
[authn-rb client](https://github.com/keratin/authn-rb) requires only a few lines:

```ruby
class ApplicationController
  private

  def logged_in?
    !! current_user
  end

  def current_user
    @current_user ||= current_account_id && User.find_by_account_id(current_account_id)
  end

  def current_account_id
    # if your client sends a cookie named "authn" containing the access token
    Keratin::AuthN.subject_from(cookies[:authn])
    # OR if your client uses localStorage and sends an Authorization header
    (request.headers['Authorization'] || '').sub(/^Bearer /, '')
  end
end
```

## Requirements

* A SQL database for long-term accounts and credentials data. (currently: PostgreSQL, MySQL, SQLite)
* A key/value database for sessions, metrics, and ephemeral data. (currently: Redis)
* A host application responsible for user data, user permissions, and email delivery. (currently:
  Go, Ruby)
* A client application responsible for user input and refreshing access tokens. (currently:
  JavaScript)
