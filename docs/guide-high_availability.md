# High Availability

AuthN is ready for a high availability deployment with multiple instances. AuthN servers coordinate
with each other to synchronize their active keys so they can verify each others' tokens.

The main requirement for this is a shared key-value storage backend. Currently only Redis has been
implemented for this role.

The second requirement is that your servers have reasonably synchronized clocks. If one server's
clock drifts too much, it may switch to a key early or continue using a key past its expiration.
