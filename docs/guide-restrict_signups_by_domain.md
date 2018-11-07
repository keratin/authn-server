# Single Domain Signups

If you wish to ensure that all users are part of the same organization, restricting signups to a
single domain and verifying email addresses is a good pattern.

## Configuration

* [USERNAME_IS_EMAIL](config.md#username_is_email)
* [EMAIL_USERNAME_DOMAINS](config.md#email_username_domains)

## Implementation

1. Configure AuthN to validate emails and domains.
2. Immediately [lock](api.md#lock-account) the account after registration.
3. Implement your email verification process.
4. [Unlock](api.md#unlock-account) the account when the email verifies.

> NOTE:
> If you also use account locking as a moderation action, be sure to control the email verification process enough that you can be confident someone will not be able to use it as a way to unlock their account unexpectedly.
