FROM alpine

WORKDIR /app
ADD dist/authn /app/authn

EXPOSE 3000
CMD ["./authn", "server", "-port", "3000"]
