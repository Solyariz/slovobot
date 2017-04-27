FROM alpine:latest

MAINTAINER Edward Muller <edward@heroku.com>

WORKDIR "/opt"

ADD .docker_build/slovobot /opt/bin/slovobot


CMD ["/opt/bin/slovobot"]

