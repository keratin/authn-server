---
title: AuthN Deployment
tags:
  - guides
  - deployment
---


## Docker

The Docker image is tagged with release versions like `keratin/authn-server:0.9.0`. Please do not
rely on the `:latest` tag, as it is not guaranteed to reflect the newest stable version.

### Daemon

You can run the AuthN Docker image as a daemon. Here's the quickest way to get it running:

```sh
# start a Redis server in the background
docker run --detach --name authn_redis redis

# then, configure and start an AuthN server on localhost:8080
docker run \
  --publish 8080:3000 \
  --link authn_redis:rd \
  -e AUTHN_URL=localhost:8080 \
  -e APP_DOMAINS=localhost \
  -e DATABASE_URL=sqlite3:db/demo.sqlite3 \
  -e REDIS_URL=redis://rd:6379/1 \
  -e SECRET_KEY_BASE=`ruby -rSecureRandom -e 'puts SecureRandom.hex(64)'` \
  -e HTTP_AUTH_USERNAME=hello \
  -e HTTP_AUTH_PASSWORD=world \
  --detach \
  --name authn_app \
  keratin/authn-server:latest \
  sh -c "./authn migrate && ./authn -port 3000 server"
```

> NOTE:
> This AuthN daemon uses a random `SECRET_KEY_BASE`, which will invalidate old sessions every time
> it starts up. Please review the configuration options to suit your environment before depending
> on this command.

### Compose

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
    image: keratin/authn-server:0.9.0
    ports:
      - "8765:3000"
    environment:
      - DATABASE_URL=mysql2://root@db:3306/authn
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
