FROM alpine:3.5

RUN apk add --no-cache ca-certificates openssh-client

# Install Docker client
RUN apk add --no-cache curl \
    && DOWNLOAD_FILE=$(curl -s --location https://download.docker.com/linux/static/stable/x86_64 | sed -n 's/.*href="\([^"]*\).*/\1/p' | grep docker- | grep .tgz | tail -n 1) \
	  && curl -L -o /$DOWNLOAD_FILE https://download.docker.com/linux/static/stable/x86_64/$DOWNLOAD_FILE \
    && tar -xz -C / -f /$DOWNLOAD_FILE \
    && mv /docker/docker /usr/bin \
    && rm -rf /$DOWNLOAD_FILE /docker \
    && apk del curl \
    && docker --version

COPY rancher /usr/bin/
RUN rancher --version
WORKDIR /mnt
ENTRYPOINT ["rancher"]
CMD  ["--help"]
