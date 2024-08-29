# Use multi-stage build
FROM golang:1.23-alpine3.20 as build

WORKDIR /go/src/autu
ADD . .


RUN go mod download

RUN go install github.com/securego/gosec/v2/cmd/gosec@latest

RUN go build -o /go/main ./cmd/api
RUN go test ./...
#RUN gosec ./...

FROM gcr.io/distroless/static
WORKDIR /go/
#COPY --chmod=0755 images/ ./images/
#COPY --chmod=0644 ./static/ /go/static/
COPY --chmod=0755 --from=build /go/main .
COPY ./ .

HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/healthz || exit 1

CMD ["./main"]
