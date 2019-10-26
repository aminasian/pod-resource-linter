FROM golang:1.13-alpine3.10

RUN mkdir /app

ADD ./pod-resource-linter /app/

WORKDIR /app

RUN adduser -S -D -H -h /app appuser

USER appuser

EXPOSE 8080

CMD ["./pod-resource-linter"]