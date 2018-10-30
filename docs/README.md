# Introduction

AuthN is an accounts microservice. It's what you get when you treat passwords like credit cards.

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

  # if your authn-js client is configured to use localstorage then `cookies[:authn]` may
  # need to be replaced by something like `request.headers['Authorization']`
  def current_account_id
    Keratin::AuthN.subject_from(cookies[:authn])
  end

  def bearer_token
    (request.headers['Authorization'] || '').sub(/^Bearer /, '')
  end
end
```
