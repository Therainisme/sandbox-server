FROM golang:1.17.7
COPY ./* /sandbox-server/executor/
WORKDIR /sandbox-server/executor/
RUN go build -o run .

FROM ubuntu:20.04
COPY --from=0 /sandbox-server/executor/run /sandbox-server/executor/
WORKDIR /sandbox-server/executor/