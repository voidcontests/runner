FROM debian:bullseye

# installing deps
RUN apt-get update
RUN apt-get install -y gcc make libc6-dev
RUN rm -rf /var/lib/apt/lists/*

# create an isolated user
RUN useradd -m -s /bin/bash sandbox

WORKDIR /sandbox

RUN chown sandbox:sandbox /sandbox

USER sandbox

CMD ["/bin/bash"]
