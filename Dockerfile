FROM golang:alpine
WORKDIR /go/src/app
COPY . .
ENV USER=go \
    UID=1000 \
    GID=1000 \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

RUN go mod tidy && \
    go build -ldflags="-s -w" \
    -o github-pr-exporter && \
    addgroup --gid "$GID" "$USER" && \
    adduser \
    --disabled-password \
    --gecos "" \
    --home "$(pwd)" \
    --ingroup "$USER" \
    --no-create-home \
    --uid "$UID" \
    "$USER" && \
    chown "$UID":"$GID" /go/src/app/github-pr-exporter

FROM scratch
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /go/src/app/github-pr-exporter /
USER 1000
ENTRYPOINT ["/github-pr-exporter"]