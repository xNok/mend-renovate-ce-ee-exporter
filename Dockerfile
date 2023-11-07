##
# BUILD CONTAINER
##

FROM alpine:3.18 as certs

RUN \
apk add --no-cache ca-certificates

FROM golang:1.21 as build

WORKDIR /app

COPY . .
RUN make build

##
# RELEASE CONTAINER
##

FROM busybox:1.36-glibc

WORKDIR /

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/mend-renovate-ce-ee-exporter /usr/local/bin/

# Run as nobody user
USER 65534

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/mend-renovate-ce-ee-exporter"]
CMD ["run"]