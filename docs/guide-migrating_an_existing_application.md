---
title: Migrating an Existing Application
tags:
  - guides
---

# Guide: Migrating an Existing Application

Migrating to Keratin AuthN is best accomplished one user account at a time. The goal is to implement an AuthN integration side-by-side with the legacy system, then detect which system a user's account exists in and act accordingly. This creates a stable transition environment that allows you to gain confidence in the system and control the migration speed.

These instructions assume that you have a users table and a UsersController. The specific names are not important, so please adapt the instructions to your system.

## 1. Adding AuthN Implementation

* add a `users.account_id` field to your schema
* update User validations to require a password OR an account_id
* update `POST /users` endpoint to set account_id instead of password when AuthN token exists
* update `current_user` helper to find user by account_id instead of session cookie when AuthN token exists
* add a `GET /users/authn` endpoint that takes a username param and returns 200 when users.account_id exists, or 404 when it does not (if you have throttling in your system to prevent user enumeration, consider also protecting this endpoint while it exists)
* when a user attempts to log in, check `GET /users/authn` and decide whether to log them in with AuthN or the legacy system
* when a user attempts to reset their password, check `GET /users/authn` and decide whether to initiate the process through AuthN or the legacy system
* when a user is logged in and wants to change their password, check `users.account_id` and decide whether to perform the action through AuthN or the legacy system
* when locking or deleting (archiving) a user, check `users.account_id` and make sure to reflect the changes in AuthN

## 2. Transitioning New Users

Your system is now ready to appropriately route users through AuthN when `users.account_id` exists. But maybe something was overlooked, or you want to manage and monitor load carefully. The best way to proceed is by introducing AuthN users into your system at a slow pace. There's no rush here, so consider a rollout plan like:

| day | action |
| --- | ------ |
| 1 | 1% of new user signups |
| 2 | watch |
| 3 | watch |
| 4 | 10% |
| 5 | 20%, 50% |
| 6 | 100% |

During this period, you should expect that some users will attempt to sign up when they mean to log in. If their signup attempt is sent to AuthN, it might be accepted! This is because AuthN's uniqueness validations do not encompass your legacy data. Have no fear, the situation can be resolved: in your system's `POST /users` endpoint, add a condition that will check for duplicate usernames accompanied by AuthN tokens, and will then update the _existing_ user's account_id instead of creating a new user. Then, if the existing user's account should be locked, be sure to immediately update AuthN.

## 3. Transitioning Existing Users

Now that every new user has an AuthN account, it's time to start transitioning your existing user accounts.

If your legacy system uses BCrypt passwords, this is easy. You can begin looping through existing accounts, [sending them to AuthN](api.md#import-account), and storing the account_id that you get back.

However, if your legacy system does not use BCrypt passwords you have a choice:

A. Submit an [issue](https://github.com/keratin/authn) describing your situation. AuthN may be able to add support for your style of password hashing. It's important to realize that these old passwords may not be as secure, though, and after some period of time (months) you should revoke them (any active users who have logged in meanwhile will have their passwords upgraded to BCrypt seamlessly).

B. Over some long period of time (1-2 months), wait for users to log in and [import them to AuthN](api.md#import-account) while you have their raw password available. This allows your most active users to transition seamlessly. Then, once you're satisfied with the number of users that have migrated, import the remaining accounts _without passwords_ and flagged for an immediate password change. The next time these inactive accounts return, they will see a prompt telling them to reset their password by the usual process.

## 4. Removing Legacy System

Congratulations! Now that every user has an AuthN account, it's time to clean up the transition support.

* Delete the legacy signup, login, password reset, and password change systems.
* Delete the `GET /users/authn` endpoint.
* Update `POST /users` to require an AuthN token and only set account_id, not passwords
* Update user validations to only require account_ids, not passwords
* Update `current_user` helper to only parse an AuthN token, not legacy cookies
* Drop unused fields from your users table (e.g. passwords, session and reset tokens, etc)
