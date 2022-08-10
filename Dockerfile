# STEP 1 build executable binary
FROM golang:alpine as builder

COPY . /src/testapp/

WORKDIR /src/testapp/

RUN go build -o /testapp

# STEP 2 build a small image
FROM alpine

RUN adduser -D -g '' testapp

COPY --from=builder /testapp /testapp
COPY --chown=testapp data/ /data

USER testapp

EXPOSE 9000

ENTRYPOINT ["/testapp"]
