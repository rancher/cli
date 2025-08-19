FROM registry.suse.com/bci/bci-base:15.7

ARG user=cli

RUN zypper -n update && \
    zypper -n install ca-certificates openssh-clients && \
    zypper clean -a && rm -rf /tmp/* /var/tmp/* /usr/share/doc/packages/* /usr/share/doc/manual/* /var/log/*

RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user

COPY rancher /usr/bin/

WORKDIR /home/$user

USER 1000:1000

ENTRYPOINT ["rancher"]
CMD  ["--help"]
