FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /sentinel

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
    -X github.com/jaxxstorm/sentinel/internal/constants.TagName=${TAG_NAME} \
    -X github.com/jaxxstorm/sentinel/internal/constants.BuildTimestamp=${BUILD_TIMESTAMP} \
    -X github.com/jaxxstorm/sentinel/internal/constants.CommitHash=${COMMIT_HASH}" ./cmd/sentinel

FROM scratch

WORKDIR /sentinel

COPY --from=builder /sentinel/sentinel .

ENV SENTINEL_CONFIG_PATH=/sentinel/config.yaml

ENV PATH=$PATH:/sentinel

ENTRYPOINT ["sentinel"]
CMD ["run"]
