FROM alpine:3.13
RUN apk add --no-cache ca-certificates openssh-client curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod u+x kubectl && mv kubectl /bin/kubectl
COPY rancher /usr/bin/
WORKDIR /mnt
ENTRYPOINT ["rancher"]
CMD  ["--help"]
