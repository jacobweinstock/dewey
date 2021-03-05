FROM golang:1.15 as builder

WORKDIR /code
COPY go.mod go.sum /code/
RUN go mod download

COPY . /code
RUN make build

FROM scratch

COPY --from=builder /code/bin/dewey-linux-amd64 /dewey-linux-amd64

ENTRYPOINT ["/dewey-linux-amd64"]