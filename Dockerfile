FROM ubuntu:latest

RUN apt update && \
    apt install -y \
    	ssh \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /home/ubuntu

COPY ./bin/setup /usr/local/bin/setup

CMD ["/bin/bash"]
