# Deployment

AuthN may be deployed to suit your needs. Currently it is provided in a Docker image, with prebuilt
binaries in GitHub Releases for Linux, MacOS, and Windows.

## Dependencies

AuthN requires:

* A SQL database (currently supports PostgreSQL, MySQL, & Sqlite3) ([submit a request](https://github.com/keratin/authn-server/issues))
* A Redis server for session tokens, ephemeral data, and activity metrics.
* Network routing from your application's clients.
* Network routing to/from your application, for secure back-channel API communication.

## Deployment Strategy

1. Provision a server (or decide to colocate it on an existing server)
2. Deploy the code
3. Set environment variables to configure the database and other settings
4. Run migrations
5. Send traffic!

## Maximum Security

Ensure that all communication to AuthN happens with SSL.

Configure [`PUBLIC_PORT`](config#public_port) and send all public traffic there.

Give AuthN its own dedicated SQL and Redis databases and be sure that all database backups are
strongly encrypted at rest. The credentials and accounts data encapsulated by AuthN should not be
necessary for data warehousing or business intelligence, so try to minimize their exposure.

## Configuration

* [PORT](config.md#port)
* [PUBLIC_PORT](config.md#public_port)
* [PROXIED](config.md#proxied)

## Related Guides

* [High Availability](guide-high_availability.md)
* [Deploying with Docker](guide-deploying_with_docker.md)
* [Integrating AuthN with an API Gateway](guide-integrating_authn_with_an_api_gateway.md)
