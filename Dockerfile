FROM apache/answer AS answer-builder

FROM golang:1.23-alpine AS golang-builder

COPY --from=answer-builder /usr/bin/answer /usr/bin/answer

RUN apk --no-cache add \
    build-base git bash nodejs npm go && \
    npm install -g pnpm@10.7.0

RUN answer build \
    --with github.com/apache/answer-plugins/connector-basic \
    --with github.com/neermilov/apache-answer-telegram-connector@518d7f6a5e3155bcb2aef5206c159870b6efe310 \
    --output /usr/bin/new_answer

FROM alpine
LABEL maintainer="linkinstar@apache.org"

ARG TIMEZONE
ENV TIMEZONE=${TIMEZONE:-"Asia/Shanghai"}

RUN apk update \
    && apk --no-cache add \
        bash \
        ca-certificates \
        curl \
        dumb-init \
        gettext \
        openssh \
        sqlite \
        gnupg \
        tzdata \
    && ln -sf /usr/share/zoneinfo/${TIMEZONE} /etc/localtime \
    && echo "${TIMEZONE}" > /etc/timezone

COPY --from=golang-builder /usr/bin/new_answer /usr/bin/answer
COPY --from=answer-builder /data /data
COPY --from=answer-builder /entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh

ENV TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-""}
ENV TELEGRAM_BOT_USERNAME=${TELEGRAM_BOT_USERNAME:-""}
ENV TELEGRAM_REDIRECT_PATH=${TELEGRAM_REDIRECT_PATH:-""}

VOLUME /data
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]

