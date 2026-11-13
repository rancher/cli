ARG BUILD_IMAGE=registry.suse.com/bci/bci-micro:16.0
ARG BASE_IMAGE=registry.suse.com/bci/bci-nano:16.0

FROM ${BUILD_IMAGE} AS builder

ARG user=cli

# bci-nano has no shell or package manager: install the runtime packages into
# a separate root and prepare user files and home directory here.
RUN zypper -n update && \
    zypper -n --installroot /stage --gpg-auto-import-keys install --no-recommends ca-certificates openssh-clients && \
    zypper -n --installroot /stage clean -a && \
    rm -rf /stage/var/cache /stage/var/log/* /stage/tmp/* /stage/var/tmp/* \
           /stage/usr/share/doc/packages/* /stage/usr/share/doc/manual/* && \
    zypper clean -a && rm -rf /tmp/* /var/tmp/* /var/log/*

RUN mkdir -p /stage/etc /stage/home/$user && \
    echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /stage/etc/passwd && \
    echo "$user:x:1000:" >> /stage/etc/group && \
    chown -R 1000:1000 /stage/home/$user

FROM ${BASE_IMAGE}

ARG user=cli

COPY --from=builder /stage/ /
COPY rancher /usr/bin/

WORKDIR /home/$user

USER 1000:1000

ENTRYPOINT ["rancher"]
CMD  ["--help"]
