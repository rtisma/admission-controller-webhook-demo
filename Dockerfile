ARG GO_VERSION=1.15
FROM golang:${GO_VERSION}-alpine AS build

COPY ./cmd /srv
COPY ./init-build.sh /srv
WORKDIR /srv

RUN apk add --no-cache git \
    && chmod +x ./init-build.sh \
	&& ./init-build.sh


env CGO_ENABLED=0 
env GOOS=linux 
RUN cd /srv/webhook-server \
	&& go build -ldflags="-s -w" -o webhook-server *.go


###############################################################

FROM golang:${GO_VERSION}-alpine

ENV APP_USER appuser
ENV APP_UID 9999
ENV APP_GID 9999
ENV APP_HOME /app

COPY --from=build /srv/webhook-server/webhook-server /tmp/webhook-server

RUN addgroup -S -g $APP_GID $APP_USER  \
	&& adduser -S -u $APP_UID -G $APP_USER $APP_USER  \
	&& mkdir -p $APP_HOME \
	&& mv /tmp/webhook-server $APP_HOME/webhook-server \
	&& chown -R $APP_UID:$APP_GID $APP_HOME

WORKDIR $APP_HOME

USER $APP_UID

CMD ["/app/webhook-server"]


