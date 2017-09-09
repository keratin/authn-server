---
title: Restrict Signups by Invitation
tags:
  - guides
---


# Guide: Restrict Signups by Invitation

If you wish to only permit signups for users with an invitation, you should take the following steps:

1. Create a private or restricted signup page for folks with a known email address or unique invitation code.
2. Integrate Keratin AuthN as normal into the signup process.
3. After signup, when you have created a user in your system and associated it with a Keratin AuthN account, determine if the user was invited or is trying to sneak past. You can do this by checking the email address or verifying the invitation code.
4. If a user is trying to sneak past your invitation system, automatically [lock](api.md#lock-account) their Keratin AuthN account and encourage them to come back later when you have launched. Remember that this is a failsafe, and not your intended or primary UX. This shouldn't happen.

When it's time to fully launch your application, don't forget to [unlock](api.md#unlock-account) any users that tried to sneak past registration!
