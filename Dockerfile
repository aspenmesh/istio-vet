FROM golang:1.8 as builder
WORKDIR /go/src/github.com/aspenmesh/istio-vet

RUN apt-get update \
 && apt-get dist-upgrade -y \
 && apt-get install -y \
    unzip \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

RUN curl -s -L \
    https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 \
    > /go/bin/dep \
 && echo "322152b8b50b26e5e3a7f6ebaeb75d9c11a747e64bbfd0d8bb1f4d89a031c2b5 /go/bin/dep" | sha256sum -c - \
 && chmod +x /go/bin/dep

RUN go get github.com/golang/protobuf/protoc-gen-go

RUN curl -L -O \
    https://github.com/google/protobuf/releases/download/v3.4.0/protoc-3.4.0-linux-x86_64.zip \
 && echo 'e4b51de1b75813e62d6ecdde582efa798586e09b5beaebfb866ae7c9eaadace4 protoc-3.4.0-linux-x86_64.zip' | sha256sum -c - \
 && mkdir -p /usr/local \
 && unzip protoc-3.4.0-linux-x86_64.zip -d /usr/local

# Install the locked go deps into vendor
COPY Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only

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
