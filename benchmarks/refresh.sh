#!/bin/sh

AUTHN=http://localhost:8080
ORIGIN=http://localhost:3000
RAND=$(openssl rand -hex 5)

SESSION=`
  curl -s -D - -o /dev/null \
    --header "Origin: $ORIGIN" \
    --header "Content-Type: application/x-www-form-urlencoded" \
    --data "username=$RAND&password=$RAND" \
    $AUTHN/accounts \
    | grep Set-Cookie \
    | sed -e 's/Set-Cookie: //' -e 's/; HttpOnly//' \
    | sed -e 's/[[:space:]]*$//'
`

wrk2 --timeout 2s \
  -t1 -c1 -d20 -R25 \
  -H "Origin: $ORIGIN" \
  -H "Cookie: $SESSION" \
  $AUTHN/session/refresh
