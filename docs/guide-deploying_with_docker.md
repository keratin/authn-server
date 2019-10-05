# Docker

The Docker image is tagged with release versions like `keratin/authn-server:0.9.0`. Please do not
rely on the `:latest` tag, as it is not guaranteed to reflect the newest stable version.

## Daemon

You can run the AuthN Docker image as a daemon. Here's the quickest way to get it running with
minimal dependencies:

```sh
docker run -it --rm \
  --publish 8080:3000 \
  -e AUTHN_URL=http://localhost:8080 \
  -e APP_DOMAINS=localhost \
  -e DATABASE_URL=sqlite3://:memory:?mode=memory\&cache=shared \
  -e SECRET_KEY_BASE=changeme \
  -e HTTP_AUTH_USERNAME=hello \
  -e HTTP_AUTH_PASSWORD=world \
  --name authn_app \
  keratin/authn-server:latest \
  sh -c "./authn migrate && ./authn server"
```

## Compose

If your application development is also Dockerized, you should consider a docker-compose.yml that
will coordinate all of its service dependencies. This is your development version of ECS or
Kubernetes.

Example:

```yml
version: '2'
services:
  db:
    image: mysql:5.7
    ports:
      - "3306:3306"
    environment:
      - MYSQL_ROOT_PASSWORD
      - MYSQL_DATABASE
      - MYSQL_ALLOW_EMPTY_PASSWORD=yes

  redis:
    image: redis

  authn:
    image: keratin/authn-server:1.0.0
    ports:
      - "8765:3000"
    environment:
      - DATABASE_URL=mysql://root@db:3306/authn
      - REDIS_URL=redis://redis:6379/0
      - AUTHN_URL=http://authn:3000
      - APP_DOMAINS=localhost
      - SECRET_KEY_BASE
    depends_on:
      - redis
      - db

  app:
    # ...
    depends_on:
      - authn
```

> NOTE:
> AUTHN_URL must be reachable by the app. Here we use the internal Docker connection. The exposed
> URL for frontend clients is http://localhost:8765, as per the authn ports mapping.
