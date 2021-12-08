FROM alpine:3.12.0 AS setu-checker
LABEL maintainer="sayuri556677@gmail.com"
RUN apk update && \
    apk add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    rm -rf /var/cache/apk/*
ENV SETU_PATHS ""
ENV SETU_AK ""
ENV SETU_SK ""
ENV SETU_FILE_TYPE ""

ENV SETU_FAIL_PATH ""
ENV SETU_NO_CHECK_PATH ""
ENV SETU_NO_H_PATH ""
ENV SETU_NORMAL_H_PATH ""
ENV SETU_ANIME_H_PATH ""
ENV SETU_SM_H_PATH ""
ENV SETU_LOW_H_PATH ""
ENV SETU_LOLI_H_PATH ""
ENV SETU_ART_H_PATH ""
ENV SETU_TOYS_H_PATH ""
ENV SETU_MEN_SEXY_H_PATH ""
ENV SETU_MEN_BARE_H_PATH ""
ENV SETU_NORMAL_SEXY_H_PATH ""
ENV SETU_ANIME_SEXY_H_PATH ""
ENV SETU_PREGNANT_H_PATH ""
ENV SETU_SPECIAL_H_PATH ""
ENV SETU_HIPS_H_PATH ""
ENV SETU_FEET_H_PATH ""
ENV SETU_CROTCH_H_PATH ""
ENV SETU_INTIMATE_H_PATH ""
ENV SETU_ANIME_INTIMATE_H_PATH ""

ADD setu-checker /setu-checker
ADD config.yaml /config.yaml
CMD ["/setu-checker"]
