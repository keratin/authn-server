# Notification Signature Validation

For notifications back to your app, AuthN can use an optional signing key to sign the message.  This is
useful for verifying that the message came from AuthN and not a malicious third party.  The signing key
should be a hex-encoded string of at least 32 bytes for use with sha256 HMAC.  Signatures will be hex-encoded
from AuthN and sent in the `X-Authn-Notification-Signature` header.

## Configuration

* [APP_SIGNING_KEY](config.md#app_signing_key)

## Implementation

### Notification Receiver

In short, the receiver should:
- encode the form parameters received as a query string with sorted keys
- calculate the HMAC of the query string using the signing key
- hex-encode the HMAC for comparison

As an example, a rails application may want to send a password reset email to a user.  The request can be
validated like so before sending:

```ruby
class AuthnController < ApplicationController
  def password_reset
    sig = request.headers["X-Authn-Notification-Signature"]
    hex_secret_key = ENV["APP_SIGNING_KEY"]
    secret_key = [hex_secret_key].pack("H*")
    payload = params.sort.to_h.map { |key, value| "#{key}=#{value}" }.join("&")
      
    # calculate an HMAC of the payload with hex-encoded output
    digest = OpenSSL::Digest.new("sha256")
    hmac = OpenSSL::HMAC.hexdigest(digest, secret_key, payload)
      
    # Compare the signatures
    unless hmac == sig
      head :unauthorized
      return
    end
    
    # Process the notification
    @user = User.find_by_account_id(params[:account_id])
    AccountMailer.password_reset(@user, params[:token]).deliver_later
    head :ok
  end
end
