# Change Password

Your users will sometimes want to change their passwords while they are logged in. Easy!

## Configuration

* [PASSWORD_POLICY_SCORE](#password_policy_score)

## Implementation

### Frontend

* Create a form where a logged-in user may enter their current and new passwords with an 
  optional TOTP MFA code (required if the user has completed MFA onboarding with their authenticator app).
* Submit the current and new passwords with the MFA code to AuthN.

## Related Guides

* [Displaying a password strength indicator](guide-displaying_a_password_strength_meter.md)
