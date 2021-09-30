FROM golang:1.16 as builder

WORKDIR /builder

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN make build

# ---------------

FROM ubuntu:latest 

RUN apt update && apt install ca-certificates tzdata -y && rm -rf /var/lib/{apt,dpkg,cache,log}/

WORKDIR /bin

COPY --from=builder /builder/bin/ataas /bin/ataas-bin

CMD ["/bin/ataas-bin"]