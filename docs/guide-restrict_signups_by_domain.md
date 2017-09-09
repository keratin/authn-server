---
title: Restrict Signups by Domain
tags:
  - guides
---

# Guide: Restrict Signups by Domain

If you wish to ensure that all users are part of the same organization, restricting signups to a single domain and verifying email addresses is a good pattern.

Here's how:

1. Configure AuthN to [validate usernames as emails](config.md#username_is_email) and [validate the domain](config.md#email_username_domains).
2. Immediately [lock](api.md#lock-account) the account after registration.
3. Implement your email verification process.
4. [Unlock](api.md#unlock-account) the account when the email verifies.*

* NOTE: if you also use account locking as a moderation action, be sure to control the email verification process enough that you can be confident someone will not be able to use it as a way to unlock their account unexpectedly.
