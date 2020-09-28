FROM golang:1.15.2
WORKDIR /go/src/atlantis-proxy/
COPY *.go .
#RUN go get -u ./...
RUN go get -u github.com/gorilla/mux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o atlantis-proxy .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /atlantis-proxy/
COPY --from=0 /go/src/atlantis-proxy/atlantis-proxy .
CMD ["./atlantis-proxy"]