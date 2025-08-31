ARG GOLANG_VERSION

#hadolint ignore=DL3006
FROM golang:${GOLANG_VERSION}-alpine AS builder

WORKDIR /src/

COPY . /src/

RUN CGO_ENABLED=0 go build -o /bin/stashly main.go

#hadolint ignore=DL3006
FROM alpine

#hadolint ignore=DL3018
RUN apk add --no-cache postgresql-client

COPY --from=builder /bin/stashly /bin/stashly

ENTRYPOINT ["/bin/stashly"]
