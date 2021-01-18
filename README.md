<!-- markdownlint-disable MD039 MD041 -->
[![Go Report Card](https://goreportcard.com/badge/github.com/stormforger/testapp)](https://goreportcard.com/report/github.com/stormforger/testapp)

<!-- markdownlint-enable MD039 MD041 -->

# StormForger Test App

This repository contains the code we run for testing purposes on [testapp.loadtest.party](http://testapp.loadtest.party). The endpoint can also be reached via TLS.

## Usage

```console
docker run --rm -p 8080:8080 -p 8443:8443 stormforger/testapp
```

* you can configure the listen port via the `PORT` and `TLS_PORT` env variables

## Endpoints

* `/demo`: Used for demos
  * [`/demo/login`](http://testapp.loadtest.party/demo/login): Has a 5% change to delay the JSON response by 250-350ms
  * [`/demo/search`](http://testapp.loadtest.party/demo/search): Will fail if query parameters are present (HTTP 400 response and different JSON response body)
* [`/data`](http://testapp.loadtest.party/data): Collection of static responses in different formats (HTML, JSON, XML)
* [`/respond-with/bytes?size=SIZE`](http://testapp.loadtest.party/respond-with/bytes?sizes=1024): Will respond with `SIZE` random bytes
* [`/do-not-respond`](http://testapp.loadtest.party:9001/do-not-respond): Will read the request and then close the connection without sending any response

* [`/`](http://testapp.loadtest.party/): All other requests will be responded to as an echo server (replying with the seen request, including the body if it is below 10kb in size).

  * If a `location` query parameter is provided to the echo endpoint, the response will contain the value of this parameter in the `Location` header
  * If a `status` query parameter is provided to the echo endpoint, the response will use the value of this parameter for the response status code. If none is provided, the response will always be `200`.

## Middlewares

* delay: All routes support a generic `delay` query parameter which specifies the number of milliseconds that the request should be artificially hold before processing
* compress: The gorillatoolkit compression handlers supports gzip encoding responses, if the correct http headers are specified

## Example

```terminal
curl -d '{"hello": "world"}' \
  -H "Content-Type: application/json" \
  'http://testapp.loadtest.party/say/hello/?foo=bar'
POST /say/hello/?foo=bar HTTP/1.1
Host: testapp.loadtest.party
Accept: */*
Content-Length: 18
Content-Type: application/json
User-Agent: curl/7.54.0

{"hello": "world"}
```

# X.509 & EST Endpoints (Enrollement over Secure Transport)

**NOTE** that the certificate material used by testapp is for testing purposes only!

* `/x509/inspect`: Can be used with a Client TLS certificate. The response will be JSON, containing the subject of the client certificate. All client certificates will be accepted.

[EST/RFC7030](https://tools.ietf.org/html/rfc7030) Endpoints:

* `/.well-known/est/cacerts`: Will return the current CA certificates in use. Note that this is a just a test certificate. See [RFC7030 4.1](https://tools.ietf.org/html/rfc7030#section-4.1) for details.

You can use OpenSSL to convert the response into PEM:

```terminal
curl https://testapp.loadtest.party/.well-known/est/cacerts | base64 -D | openssl pkcs7 -inform DER -print_certs
```

* `/.well-known/est/simpleenroll`: If you POST a base64 encoded PKCS10 to this endpoint, you will get a base64 encoded PKCS7 response. In contrast to RFC7030 no authentication is required.

You can generate a new private key and a CSR using `openssl` and `base64` (as RFC7030 requires base64 encoded PKCS10):

```terminal
openssl req -new -newkey rsa:2048 -nodes -out tmp/client.csr.der -outform DER -keyout tmp/client.key.pem -subj "/CN=hello-world"
base64 tmp/client.csr.der > tmp/client.csr.b64
curl -k -X POST --data-binary @tmp/client.csr.b64 -o tmp/cert.p7.base64 -k https://localhost:8443/.well-known/est/simpleenroll -H'Content-Transfer-Encoding: base64'
cat tmp/cert.p7.base64 | base64 -D | openssl x509 -inform DER > tmp/client.crt.pem
```

Alternatively you can use the `client/main.go` tool to generate the CSR + private key file.

Usage:

```terminal
curl --cert ./tmp/client.crt.pem --key ./tmp/client.key.pem -k https://localhost:8443/x509/inspect
```

## Build & Release

```terminal
docker build . -t stormforger/testapp
docker push stormforger/testapp
```

### Generate Server RSA Key and Certificate

```terminal
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go --host localhost
mv cert.pem data/pki/server.cert.pem
mv key.pem data/pki/server.key.pem
```
