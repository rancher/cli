FROM golang:1.7.3
RUN go get github.com/rancher/trash
RUN go get github.com/golang/lint/golint
RUN curl -sL https://get.docker.com/builds/Linux/x86_64/docker-1.9.1 > /usr/bin/docker && \
    chmod +x /usr/bin/docker
RUN apt-get update && \
    apt-get install -y --no-install-recommends default-jre python-pip zip xz-utils rsync
RUN pip install --upgrade pip==6.0.3 tox==1.8.1 virtualenv==12.0.4
ENV PATH /go/bin:$PATH
ENV DAPPER_SOURCE /go/src/github.com/rancher/rancher-compose
ENV DAPPER_OUTPUT bin dist
ENV DAPPER_DOCKER_SOCKET true
ENV DAPPER_ENV TAG REPO CROSS
ENV TRASH_CACHE ${DAPPER_SOURCE}/.trash-cache
WORKDIR ${DAPPER_SOURCE}
ENTRYPOINT ["./scripts/entry"]
CMD ["ci"]
