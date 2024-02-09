FROM golang:1.22.0-alpine AS builder

WORKDIR /src/

COPY . /src/

RUN CGO_ENABLED=0 go build -o /bin/gopg2s3dump main.go

FROM alpine

RUN apk add --no-cache postgresql-client

COPY --from=builder /bin/gopg2s3dump /bin/gopg2s3dump

ENTRYPOINT ["/bin/gopg2s3dump"]