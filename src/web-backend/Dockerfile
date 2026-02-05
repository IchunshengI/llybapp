# Use the locally-cached Go base image to avoid pulling from Docker Hub in restricted environments.
FROM go1_23:10_03

# Dev container: Go toolchain + sshd so the backend can be accessed via SSH (see compose/web-backend/docker-compose.yml).
RUN apt-get update && apt-get install -y --no-install-recommends \
    openssh-server \
    ca-certificates \
    git \
    passwd \
  && rm -rf /var/lib/apt/lists/*

# sshd expects these directories/files to exist.
RUN mkdir -p /run/sshd \
  && ssh-keygen -A

WORKDIR /app

# Keep the image generic; the source code is mounted in via docker-compose volume.
EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]
