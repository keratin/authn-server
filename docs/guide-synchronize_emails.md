# Synchronize Emails

If your application uses emails as usernames, then you're likely to end up with email fields in both
your user profile data and your AuthN account data. The email exists in both places so that AuthN
may take care of logins while your application mailers have easy access to addresses for email
delivery.

It's important to keep these in sync. When emails change in your app, queue a background job that
will also update the user's AuthN account.

## Configuration

[`USERNAME_IS_EMAIL`](config.md#username_is_email)

## Implementation

### Backend

1. Detect changes to user profile email addresses.
2. Queue a job (or thread) to send the new email address to AuthN.
