FROM golang:1.21.4-bookworm as builder

COPY . /app

WORKDIR /app

RUN go install ./cmd/app
RUN go install ./cmd/appcli

FROM alpine as final

EXPOSE 8080
EXPOSE 8083
EXPOSE 9000

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /go/bin/app /bin/
COPY --from=builder /go/bin/appcli /bin/
COPY --from=builder /app/db.json /

# Add bash and curl for debugging.
RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates

ENTRYPOINT ["app"]
