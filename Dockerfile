FROM golang:1.20.0-alpine
WORKDIR /app
ADD . /app
RUN cd /app && go build -o app
ENTRYPOINT /app/app
