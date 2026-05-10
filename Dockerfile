FROM alpine:3.21
RUN apk add --no-cache ca-certificates
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/helm-blame /usr/local/bin/helm-blame
ENTRYPOINT ["helm-blame"]
