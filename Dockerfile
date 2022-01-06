FROM golang:1.17-alpine as builder

WORKDIR /usr/src
COPY . .
# CGO_ENABLED=0 is required for the image to run on scratch, where shared
# libraries the go runtime expects are not available
RUN CGO_ENABLED=0 go build -o /fileserve /usr/src/cmd/fileserve
RUN adduser -D brad

FROM scratch

EXPOSE 8007
WORKDIR /var/data

# Copy over the user
COPY --from=builder /etc/passwd /etc/passwd

# Copy the fileserve static binary
COPY --from=builder /fileserve /fileserve

# Caller can specify -show-hidden to see hidden files
CMD ["/fileserve", "--listen=:8007", "--data-dir=/var/data"]