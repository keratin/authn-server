---
title: Forgotten Passwords
tags:
  - guides
---

When a user forgets their password, the best solution is to send them an email containing a link
that allows them to create a new one. AuthN ensures that the forgotten password process does not
allow attackers to enumerate user accounts while provides a secure reset token to your application
for delivery.

## Configuration

* [PASSWORD_RESET_TOKEN_TTL](config.md#password_reset_token_ttl)
* [APP_PASSWORD_RESET_URL](config.md#app_password_reset_url)
* [`PASSWORD_POLICY_SCORE`](config.md#password_policy_score)

## Implementation

Frontend:

1. Create a form where the user may enter an account email.
2. Submit the email to AuthN, and display success (AuthN always claims success).

Backend:

3. Create an endpoint (secured by SSL + HTTP Basic Auth) that expects to receive `account_id` and
   `token` params
4. Use the `account_id` to find which User is requesting a reset
5. Embed the `token` param into a link back to your frontend, and send to the user.

Frontend:

6. Create a form where the user may enter a new password. This form also needs the earlier `token`,
   likely from the current URL.
7. Submit the `token` and new password to AuthN.
8. If successful, the user will be logged in.

## Related Guides

* [Displaying a password strength indicator](guide-displaying_a_password_strength_meter.md)
