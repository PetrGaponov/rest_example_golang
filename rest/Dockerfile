FROM golang:1.11 AS builder
ADD https://github.com/golang/dep/releases/download/v0.5.3/dep-linux-amd64  /usr/bin/dep
RUN chmod +x /usr/bin/dep
ENV APP_DIR=$GOPATH/src/rest/
RUN mkdir -p $APP_DIR
COPY ./* $APP_DIR
WORKDIR $APP_DIR
RUN mkdir -p /opt/rest

###
# Copy the code from the host and compile it
#WORKDIR $GOPATH/src/github.com/username/repo
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /opt/rest/rest *.go
COPY .env  /opt/rest/

#FROM scratch
FROM alpine:3.9 as prod_img
RUN apk update && apk add ca-certificates && apk add tzdata
RUN mkdir -p /opt/rest/
RUN cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime
RUN echo "Europe/Moscow" >  /etc/timezone
COPY --from=builder /opt/rest/* /opt/rest/
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.5.0/wait /wait
RUN chmod +x /wait
## Launch the wait tool and then your application
WORKDIR /opt/rest/
CMD /wait && /opt/rest/rest
##WORKDIR /opt/rest/
##ENTRYPOINT ["/opt/rest/rest"]
