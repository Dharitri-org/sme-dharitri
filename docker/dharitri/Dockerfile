FROM golang:1.13.6 as builder
MAINTAINER DharitriNetwork

WORKDIR /go/sme-dharitri
COPY . .
RUN GO111MODULE=on go mod vendor
#Dharitri node
WORKDIR /go/sme-dharitri/cmd/node
RUN go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
RUN cp /go/pkg/mod/github.com/!Dharitri!org/sme-core-vm-go@$(cat /go/sme-dharitri/go.mod | grep core-wasm-vm | sed 's/.* //')/wasmer/libwasmer_linux_amd64.so /lib/libwasmer_linux_amd64.so
WORKDIR /go/sme-dharitri
RUN go get github.com/Dharitri-org/sme-core-vm-go/cmd/core@$(cat /go/sme-dharitri/go.mod | grep core-wasm-vm | sed 's/.* //')
RUN go build -o ./core github.com/Dharitri-org/sme-core-vm-go/cmd/core
RUN cp /go/sme-dharitri/core /go/sme-dharitri/cmd/node/
WORKDIR /go/sme-dharitri/cmd/node
# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder "/go/sme-dharitri/cmd/node" "/go/sme-dharitri/cmd/node/"
COPY --from=builder "/lib/libwasmer_linux_amd64.so" "/lib/libwasmer_linux_amd64.so"
WORKDIR /go/sme-dharitri/cmd/node/
EXPOSE 8080
ENTRYPOINT ["/go/sme-dharitri/cmd/node/node"]
