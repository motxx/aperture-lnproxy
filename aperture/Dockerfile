FROM golang:1.22.1-alpine3.19 as builder

COPY . /app

WORKDIR /app

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

RUN apk add make
RUN make install

FROM alpine as final

EXPOSE 8080
EXPOSE 8083
EXPOSE 9000

# Copy the binaries and entrypoint from the builder image.
COPY --from=builder /go/bin/aperture /bin/

# Add bash and curl for debugging.
RUN apk add --no-cache \
    bash \
    curl \
    ca-certificates

ENTRYPOINT ["aperture"]
