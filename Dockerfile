FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY helm-blame /usr/local/bin/helm-blame
ENTRYPOINT ["helm-blame"]
