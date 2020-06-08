FROM alpine:3.12.0

RUN mkdir /app
WORKDIR /app

RUN addgroup -g 1000 service && \
    adduser -u 1000 -D -G service service

COPY --chown=1000:1000 node-relabeler /app

ENTRYPOINT ["/app/node-relabeler"]
USER service