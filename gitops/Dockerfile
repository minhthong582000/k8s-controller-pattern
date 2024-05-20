FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22 as builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN mkdir -p /go/src/github.com/minhthong582000/k8s-controller-pattern/gitops
WORKDIR /go/src/github.com/minhthong582000/k8s-controller-pattern/gitops

# Cache the download before continuing
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*")

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go test -v ./...

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /usr/bin/gitops .

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=github.com/minhthong582000/k8s-controller-pattern/gitops

WORKDIR /
COPY --from=builder /usr/bin/gitops /
USER nonroot:nonroot

ENTRYPOINT ["/gitops"]
