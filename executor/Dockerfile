FROM golang:1.17.7
COPY ./* /com.therainisme/sandbox/executor/
WORKDIR /com.therainisme/sandbox/executor/
RUN go build -o run .

FROM ubuntu:20.04
COPY --from=0 /com.therainisme/sandbox/executor/run /com.therainisme/sandbox/executor/
WORKDIR /com.therainisme/sandbox/executor/