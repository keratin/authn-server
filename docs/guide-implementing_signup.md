# Signup

AuthN takes care of account data (username, password) and leaves the remaining user profile data to
your application. This means that signup is a two-step process. If that inspires you to create a
two-step user flow, that could be nice! But you can also create a traditional one-step signup
process that submits with two API calls.

## Configuration

* [USERNAME_IS_EMAIL](config.md#username_is_email)
* [EMAIL_USERNAME_DOMAINS](config.md#email_username_domains)
* [PASSWORD_POLICY_SCORE](config.md#password_policy_score)

## Implementation

### Frontend

1. Create a form that collects a user's preferred username (email?) and password.
2. Additionally collect other fields like name, newsletter subscriptions, as needed.
3. Validate everything, especially your account fields.
4. Submit the username (email?) and password to AuthN.
5. If AuthN creates an account, the user will be logged in _without a user profile_.
6. Submit the remaining details to your application.

> NOTE:
> If a user succeeds in creating an account (step 5) but fails to create a user (step 7) then your
> frontend needs a plan for how to rerun step 6 while preserving the result of step 5. Validating
> the data client-side (step 3) is important because it will significantly reduce the odds of this
> happening.

> NOTE:
> If you are using emails as username (by setting [USERNAME_IS_EMAIL](config.md#username_is_email) to `true`) and you need the email for user profile as well then consider using either authn or your application as the source of truth. You can either:
> * Use authn as the source of truth and use [Get Account](api.md#get-account) API to get the user's email when creating the user profile.
> * Use your application as the source of truth and use the [Update Account](api.md#update) API to update the email in authn when creating the user profile.

### Backend

7. Validate and save the user's AuthN `account_id` along with your other user profile fields.

## Related Guides

* Check for available usernames in real-time
* [Displaying a password strength indicator](guide-displaying_a_password_strength_meter.md)
* [Restrict signups by email domain](guide-restrict_signups_by_domain.md)
* [Restrict signups by invitation](guide-restrict_signups_by_invitation.md)
* Improve conversion with delayed signups
