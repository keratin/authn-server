# Invitation-Only Signups

AuthN does not have invitation codes built in, but you can use account locking to ensure that only
invited users gain access to your application.

## Implementation

1. Create a private or restricted signup page for folks with a known email address or unique invitation code.
2. Integrate Keratin AuthN as normal into the signup process.
3. After signup, when you have created a user in your system and associated it with a Keratin AuthN account, determine if the user was invited or is trying to sneak past. You can do this by checking the email address or verifying the invitation code.
4. If a user is trying to sneak past your invitation system, automatically [lock](api.md#lock-account) their Keratin AuthN account and encourage them to come back later when you have launched. Remember that this is a failsafe, and not your intended or primary UX. This shouldn't happen.

> NOTE:
> When it's time to fully launch your application, don't forget to [unlock](api.md#unlock-account) any users that tried to sneak past registration!
