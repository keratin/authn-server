# Password Confirmation

Your application may have critical actions that require confirming the user's password. You can
implement this by checking the `auth_time` claim of the user's access token.

## Implementation

### Backend

Your critical actions should compare `auth_time` from the user's access token (aka session) against
the current time. If you determine that the `auth_time` is too old, then the user is not authorized.

### Frontend

When the backend indicates that a user is not authorized because their login is too old, show a
password prompt to the user. Submit that password (along with the username that you already know) to
AuthN as a normal login request.

The user should now have a recent `auth_time` that will satisfy the backend's authorization control.
