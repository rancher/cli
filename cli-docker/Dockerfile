FROM alpine:3.5
MAINTAINER "Stian Larsen <lonixx@gmail.com>"
ENV version=v0.4.1
RUN \
apk add --no-cache ca-certificates wget tar openssh-client && \
wget -O /tmp/rancher.tar.gz https://github.com/rancher/cli/releases/download/${version}/rancher-linux-amd64-${version}.tar.gz && \
tar xfz /tmp/rancher.tar.gz --strip-components=2 -C /usr/bin && \
apk del --no-cache ca-certificates wget tar
WORKDIR /mnt
ENTRYPOINT ["rancher"]
CMD  ["--help"]
 
