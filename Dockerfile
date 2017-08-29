FROM alpine

WORKDIR /app
ADD dist/authn /app/authn

EXPOSE 3000
CMD ["./authn", "-port", "3000", "server"]
