FROM golang:1.12.4-alpine3.9 AS builder

RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /src
COPY . .
RUN go build -o /out/server

FROM alpine:3.9
COPY --from=builder /out/server /server
COPY --from=builder /src/mocks.yaml /config/mocks.yaml
ENTRYPOINT [ "/server" ]
CMD ["--mocks", "/config/mocks.yaml"]
