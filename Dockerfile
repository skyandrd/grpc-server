FROM golang:1.15-alpine as builder

RUN set -ex \
  && apk add --update --no-cache tzdata zip git upx \
  && cd /usr/share/zoneinfo \
  && zip -r -0 /zoneinfo.zip .

ARG CI_COMMIT_TAG="local"
ARG UPX_LEVEL="9"
COPY . /app
RUN set -ex \
  && cd /app \
  && echo "Version: ${CI_COMMIT_TAG}" \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
     time go build -mod=vendor -a -installsuffix cgo -ldflags="-w -s -X main.Version=${CI_COMMIT_TAG}" -o /app.bin \
# Optional compression
  && if [ "${UPX_LEVEL}" = "ultra" ]; then  time upx --ultra-brute /app.bin ; \
     elif [ "${UPX_LEVEL}" = "9" ]; then  time upx -9 /app.bin ; \
     fi \
  && go version

FROM scratch
ENV ZONEINFO /zoneinfo.zip
ENV TZ Asia/Yekaterinburg
COPY --from=builder /zoneinfo.zip /
COPY --from=builder /app/internal/service/mockdata/test1.csv /mockdata/test1.csv
COPY --from=builder /app/internal/service/mockdata/test2.csv /mockdata/test2.csv
COPY --from=builder /app.bin /grpc-server

CMD ["/grpc-server"]

ARG BUILD_INFO="no info"
ENV BUILD_INFO=$BUILD_INFO