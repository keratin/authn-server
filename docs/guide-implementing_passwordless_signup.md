# Passwordless Signups

If you intend to implement [Passwordless Logins](guide-implementing_passwordless_logins.md) in your
application, you may also decide to allow passwordless signups.

This pattern does not require any support from AuthN. You may simply generate a random password on
behalf of the user during signup.

Please be sure to use a cryptographically strong source. In JavaScript this means using
[`window.crypto.getRandomValues()`](https://developer.mozilla.org/en-US/docs/Web/API/Crypto/getRandomValues)
rather than `Math.random()`.

## Related Guides

* [Passwordless Logins](guide-implementing_passwordless_logins.md)
