FROM golang:1.19.12 as builder
MAINTAINER Dharitri-org

WORKDIR /go/sme-dharitri
COPY . .
RUN GO111MODULE=on go mod vendor
# Seed node
WORKDIR /go/sme-dharitri/cmd/seednode
RUN go build

# ===== SECOND STAGE ======
FROM ubuntu:18.04
COPY --from=builder /go/sme-dharitri/cmd/seednode /go/sme-dharitri/cmd/seednode

WORKDIR /go/sme-dharitri/cmd/seednode/
EXPOSE 10000
ENTRYPOINT ["./seednode"]
