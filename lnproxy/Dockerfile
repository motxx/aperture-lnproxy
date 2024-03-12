FROM golang:1.21.4-bookworm as builder

COPY . /app

WORKDIR /app

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

RUN go install ./cmd/http-relay

FROM alpine as final

EXPOSE 4747

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /go/bin/http-relay /bin/

# Add bash and curl for debugging.
RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates

ENTRYPOINT ["http-relay"]
