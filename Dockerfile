FROM golang:1.18 as build

RUN apt-get update && apt-get -y install libsodium-dev

RUN mkdir -p /sa-key-rotation/
COPY . /sa-key-rotation/
WORKDIR /sa-key-rotation

ENV GO111MODULE=on
RUN make install
RUN make build

# Now copy it into our base image.
FROM gcr.io/distroless/base
COPY --from=build /sa-key-rotation/cmd/cmd /cmd

ENTRYPOINT ["/cmd"]