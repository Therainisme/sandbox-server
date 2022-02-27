FROM golang:1.17.7
COPY ./ /sandbox-server/
WORKDIR /sandbox-server/
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn
RUN go build -o start .

FROM ubuntu:20.04
COPY --from=0 /sandbox-server/start /sandbox-server/start
WORKDIR /sandbox-server/