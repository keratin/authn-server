FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app
ADD dist/authn-linux64 /app/authn

EXPOSE 3000
ENV PORT 3000
CMD ["./authn", "server"]
