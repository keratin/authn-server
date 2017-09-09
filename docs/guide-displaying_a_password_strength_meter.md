---
title: Displaying a Password Strength Meter
tags:
  - guides
---

# Guide: Displaying a Password Strength Meter

AuthN uses a cross-platform password strength algorithm called [zxcvbn](https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/) with a default minimum strength score of `2`. This means you can predict in real-time how a password will validate and provide the user with early feedback.

If your client is written in JavaScript, you'll want to integrate [dropbox/zxcvbn](https://github.com/dropbox/zxcvbn). Your design should indicate a password strength from 1-4, and optionally indicate when the user's password has sufficient strength to validate.
