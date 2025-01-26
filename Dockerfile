FROM debian:bullseye

# Install necessary dependencies
RUN apt-get update && apt-get install -y gcc make libc6-dev && rm -rf /var/lib/apt/lists/*

# Create an isolated user
RUN useradd -m -s /bin/bash sandbox

# Set the working directory
WORKDIR /sandbox

# Set permissions for the working directory
RUN chown sandbox:sandbox /sandbox

# Switch to the isolated user
USER sandbox

# Default command (optional, can be overridden in `docker run`)
CMD ["/bin/bash"]
