# build static binary
FROM golang:1.15.6-alpine3.12 as builder 

# hadolint ignore=DL3018
RUN apk --no-cache add  \
    ca-certificates \
    git 

WORKDIR /go/src/github.com/bots-house/share-file-bot

# download dependencies 
COPY go.mod go.sum ./
RUN go mod download 

COPY . .

ARG REVISION

# compile 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-w -s -extldflags \"-static\" -X \"main.revision=${REVISION}\"" -a \
      -tags timetzdata \
      -o /bin/share-file-bot .

# run 
FROM scratch


COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/share-file-bot /bin/share-file-bot

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/bin/share-file-bot", "-health" ]

EXPOSE 8000

ENTRYPOINT [ "/bin/share-file-bot" ]