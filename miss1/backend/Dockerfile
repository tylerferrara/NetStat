#  build the binary FIRST
#  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o back
FROM alpine
WORKDIR /go/bin
COPY ./back /go/bin/back
RUN chmod +x /go/bin/back
CMD [ "/go/bin/back" ]