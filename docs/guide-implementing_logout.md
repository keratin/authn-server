# Logout

Most of your users will probably disappear when they close the application, but sometimes they will
want to cleanly log out. The AuthN client library will take care of cleaning up its sessions (the
the access token and refresh token). Your application may want to do more.

## Implementation

### Frontend

1. Create a link that users can click.
2. Invoke AuthN's logout.
3. Clean up anything else.
4. Redirect to logged-out page.
