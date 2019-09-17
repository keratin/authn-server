# OAuth Provider

AuthN provides identity management to your application. If you want to enable third-party
applications to also authenticate your users and access your APIs, then you want to take the
additional step of becoming an OAuth Provider.

This can be accomplished without AuthN's awareness. You will need to build login and consent pages
as normal, using AuthN's APIs and session management. Then you can configure a standalone OpenID
Connect provider such as [Hydra](https://github.com/ory/hydra) to handle the new responsibilities.
