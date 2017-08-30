FROM debian

ENV DEBIAN_FRONTEND noninteractive
RUN apt-get -y update && apt-get install -y ca-certificates
COPY build/batchcli-v0.0.12-linux-amd64/batchcli /usr/local/bin/batchcli
RUN chmod +x /usr/local/bin/batchcli

WORKDIR /
ADD fail_half_the_time.sh /fail_half_the_time.sh

ENTRYPOINT ["/usr/local/bin/batchcli"]
CMD ["--cmd", "/fail_half_the_time.sh"]
