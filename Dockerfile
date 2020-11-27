FROM golang:1.14 AS builder

COPY . "/go/src/github.com/aglide100/chicken_review_webserver"
WORKDIR "/go/src/github.com/aglide100/chicken_review_webserver/"

RUN mkdir -p /opt/bin/webd/
RUN mkdir -p /var/lib/webd/

#COPY go.mod .
#COPY go.sum .

RUN go mod vendor \
    && go get -u -d -v ./...
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' /opt/bin/webd/webd ./cmd/webd

RUN go build -mod=mod -o /opt/bin/webd/webd ./cmd/webd


RUN cp -r ui /var/lib/webd/
RUN cp -r pkg /var/lib/webddo


FROM debian:stretch-slim AS runtime
COPY --from=builder /opt/bin/webd/webd /opt/bin/webd/webd
COPY --from=builder /var/lib/webd /var/lib/webd

# For Using ssl certification and reqeust https api 
# add ca list in Docker container
# Change mirro to  http://mirror.kakao.com/debian/
#RUN cd /etc/apt && \
#       sed -i 's/archive.ubuntu.com/ftp.kr.debian.org/debian/' sources.list 
       #sed -i 's/archive.ubuntu.com/http://mirror.kakao.com/debian/' sources.list.d/proposed.list
RUN apt update && apt install -y ca-certificates
RUN chmod 644 /usr/local/share/ca-certificates && update-ca-certificates

WORKDIR /var/lib/webd
CMD [ "/opt/bin/webd/webd" ]
