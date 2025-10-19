FROM debian:bullseye

# installing dependencies
RUN apt-get update && \
    apt-get install -y g++ make libc6-dev python3 python3-pip && \
    rm -rf /var/lib/apt/lists/*

# create an isolated user
RUN useradd -m -s /bin/bash sandbox

WORKDIR /sandbox

RUN chown sandbox:sandbox /sandbox

USER sandbox

CMD ["/bin/bash"]
