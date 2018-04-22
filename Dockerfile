FROM golang:1.10.1-stretch
RUN go get github.com/armon/go-socks5
RUN mkdir -p src/telerus
COPY . src/telerus
RUN cd src/telerus && go build && go install
CMD ["./bin/telerus"]