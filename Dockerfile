FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /sentinel

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY ./internal ./internal
COPY ./cmd ./cmd

ARG TAG_NAME
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH
ARG CGO_ENABLED=0

# -s -w can be added to strip debug symbols and reduce binary size
RUN CGO_ENABLED=${CGO_ENABLED} go build -ldflags "\
    -X github.com/jaxxstorm/sentinel/internal/version.TagName=${TAG_NAME} \
    -X github.com/jaxxstorm/sentinel/internal/version.BuildTimestamp=${BUILD_TIMESTAMP} \
    -X github.com/jaxxstorm/sentinel/internal/version.CommitHash=${COMMIT_HASH}" ./cmd/sentinel

FROM scratch

WORKDIR /sentinel

COPY --from=builder /sentinel/sentinel .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV PATH=$PATH:/sentinel

ENTRYPOINT ["sentinel"]
CMD ["run"]
