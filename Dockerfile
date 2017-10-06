FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app
ADD dist/authn /app/authn

EXPOSE 3000
ENV PORT 3000
CMD ["./authn", "server"]
