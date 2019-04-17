FROM golang:1.10 as builder
WORKDIR /go/src/github.com/aspenmesh/istio-vet

RUN apt-get update \
 && apt-get dist-upgrade -y \
 && apt-get install -y \
    unzip \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

RUN curl -s -L \
    https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 \
    > /go/bin/dep \
 && echo "287b08291e14f1fae8ba44374b26a2b12eb941af3497ed0ca649253e21ba2f83 /go/bin/dep" | sha256sum -c - \
 && chmod +x /go/bin/dep

RUN curl -L -O \
    https://github.com/google/protobuf/releases/download/v3.4.0/protoc-3.4.0-linux-x86_64.zip \
 && echo 'e4b51de1b75813e62d6ecdde582efa798586e09b5beaebfb866ae7c9eaadace4 protoc-3.4.0-linux-x86_64.zip' | sha256sum -c - \
 && mkdir -p /usr/local \
 && unzip protoc-3.4.0-linux-x86_64.zip -d /usr/local

# Install the locked go deps into vendor
COPY Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only

RUN go install ./vendor/github.com/golang/protobuf/protoc-gen-go

COPY Makefile Makefile

COPY . .

RUN make clean
RUN make

FROM debian:stretch
WORKDIR /app

RUN apt-get update \
 && apt-get dist-upgrade -y \
 && apt-get install -y \
    curl \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/vet /usr/local/bin

CMD ["vet"]
