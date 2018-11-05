# This Dockerfile is meant to be used by the CI system, for standalone builds, use Dockerfile-standalone
FROM alpine
RUN apk update &&\
    apk add --no-cache ca-certificates &&\
    rm -rf /var/cache/apk/*
COPY release/nagios-svc /bin/
ENTRYPOINT [ "/bin/nagios-svc" ]