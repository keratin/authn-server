---
title: Integrating AuthN with an API Gateway
tags:
  - guides
---

# Guide: Integrating AuthN with an API Gateway

An API Gateway is sometimes used in a services architecture to control access and route requests to backend services. The concept is broad enough that it may be as simple as a request router -- managing service discovery so your clients don't have to -- or may take additional responsibilities like authentication, authorization, and throttling.

Keratin AuthN can work in either extreme, but I do recommend handling authentication in the gateway. This means unpacking the JWT (JSON Web Token, used by Keratin for authentication) into a simpler user ID that you pass along to backend services with some kind of "on behalf of user" parameter or header. This will take care of a requirement that your services likely share, and set you up to also handle things like per-user throttling.

Alternately, you can always pass the JWT directly through to your backend services. They will each need to integrate a Keratin AuthN backend library (e.g. `keratin/authn-rb`) to securely validate the JWT, and then perform whatever other manner of authorization logic you may have.

Regardless of where you resolve authentication issues, be sure to also:

* proxy through any [user-facing (aka public) endpoints](api.md) to the AuthN service, as if it were one of your own.
* expose login, signup, logout, and other user-facing features in your frontend client, as always.
* integrate AuthN account locking, unlocking, archival, and other private endpoints through your users service.
