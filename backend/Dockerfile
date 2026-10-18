FROM golang:1.24-bookworm AS build
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /api /api
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/api"]
