FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app
ADD dist/linux/amd64/authn-server /app/authn

EXPOSE 3000
ENV PORT 3000
CMD ["./authn", "server"]
