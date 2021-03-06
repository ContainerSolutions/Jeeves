# builder image
FROM golang:1.15-buster as builder


RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -o jeeves ./cmd/jeeves

# generate clean, final image for end users
FROM gcr.io/distroless/base-debian10
USER nonroot
COPY --from=builder /build/jeeves /
CMD ["/jeeves"]
