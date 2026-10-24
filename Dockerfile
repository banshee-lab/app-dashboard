FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /appdash .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /appdash /appdash
COPY apps.yaml /config/apps.yaml
ENV APPS_FILE=/config/apps.yaml PORT=8080
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/appdash"]
