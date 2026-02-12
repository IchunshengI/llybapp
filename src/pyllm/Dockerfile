FROM python:3.10-slim

# 安装系统依赖
RUN sed -i 's|http://deb.debian.org/debian|http://mirrors.aliyun.com/debian|g' /etc/apt/sources.list \
  && apt-get update && apt-get install -y --no-install-recommends \
    wget \
    curl \
    openssh-server \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# 配置 SSH 服务
RUN mkdir -p /run/sshd \
  && ssh-keygen -A

WORKDIR /llm

# 暴露SSH 端口
EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]