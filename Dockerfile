FROM golang:1.19-alpine
RUN apk add build-base
WORKDIR /app
RUN mkdir cmd
COPY . .
RUN go build -tags musl -o cmd

FROM alpine:3.17
WORKDIR /app
COPY --from=0 /app/cmd .
EXPOSE 443
CMD ["./telegram"]