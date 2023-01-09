FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

ADD . .
RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin ./cmd/app/

FROM scratch

COPY --from=builder ./app/bin .

EXPOSE 8082

ENTRYPOINT [ "/bin" ]
