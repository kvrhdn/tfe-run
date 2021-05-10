FROM golang:1.16.4-alpine
WORKDIR /app
ADD . /app
RUN cd /app && go build -o app
ENTRYPOINT /app/app
