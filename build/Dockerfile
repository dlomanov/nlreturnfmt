FROM golang:1.24-alpine

ENV GOPROXY=https://proxy.golang.org,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Set build-time variables
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o nlreturnfmt ./cmd/nlreturnfmt

ENTRYPOINT ["./nlreturnfmt"]
CMD ["--help"]
