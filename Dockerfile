# build static binary
FROM golang:1.16.4-alpine3.12 as builder 

# hadolint ignore=DL3018
RUN apk --no-cache add  \
    ca-certificates \
    git 

WORKDIR /go/src/github.com/bots-house/share-file-bot

# download dependencies 
COPY go.mod go.sum ./
RUN go mod download 

COPY . .

# git tag 
ARG BUILD_VERSION

# git commit sha
ARG BUILD_REF

# build time 
ARG BUILD_TIME

# compile 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-w -s -extldflags \"-static\" -X \"main.buildVersion=${BUILD_VERSION}\" -X \"main.buildRef=${BUILD_REF}\" -X \"main.buildTime=${BUILD_TIME}\"" \
      -a \
      -tags timetzdata \
      -o /bin/share-file-bot .


# run 
FROM scratch


COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/share-file-bot /bin/share-file-bot

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/bin/share-file-bot", "-health" ]

EXPOSE 8000

# git tag 
ARG BULD_VERSION

# git commit sha
ARG BUILD_REF

# build time 
ARG BUILD_TIME

# Reference: http://label-schema.org/rc1/
LABEL org.label-schema.schema-version="1.0" \
      org.label-schema.build-date=${BUILD_TIME} \
      org.label-schema.name="share-file-bot" \
      org.label-schema.description="Share files using Telegram as Cloud" \
      org.label-schema.url="https://t.me/share_file_bot" \ 
      org.label-schema.vcs-url="https://github.com/bots-house/share-file-bot" \
      org.label-schema.vcs-ref=${BUILD_REF} \
      org.label-schema.vendor="Bots House" \
      org.label-schema.version=${BUILD_VERSION} \ 
      docker.cmd.help="docker run --rm docker.pkg.github.com/bots-house/share-file-bot/share-file-bot --help"


ENTRYPOINT [ "/bin/share-file-bot" ]