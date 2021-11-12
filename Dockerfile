from golang:1.17 as builder

WORKDIR $GOPATH/src/github.com/shanehowearth/weather
ADD . $GOPATH/src/github.com/shanehowearth/weather

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /go/bin/weather cmd/main.go

from scratch

WORKDIR /root/
COPY --from=0 /go/bin/weather .

ENTRYPOINT ["./weather"]
