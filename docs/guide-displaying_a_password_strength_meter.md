# Password Strength Meter

AuthN uses a cross-platform password strength algorithm called [zxcvbn](https://blogs.dropbox.com/tech/2012/04/zxcvbn-realistic-password-strength-estimation/) with a default minimum strength score of `2`. This means you can predict in real-time how a password will validate and provide the user with early feedback.

If your client is written in JavaScript, you might consider integrating [dropbox/zxcvbn](https://github.com/dropbox/zxcvbn). The backend's password scoring is compatible with v4.4.2, specifically. Your design should indicate a password strength from 1-4, and optionally indicate when the user's password has sufficient strength to validate.

If you don't want to include this library in your web app, you can also use the [password score](api.md#password-score) endpoint to calculate the score for you. Depending on the size of the password chosen by the user, scoring might take some time - some tests on my machine yielded 100 ~ 300 ms. It will be a good idea to debounce the requests for scoring in the front-end while the user is still typing. 