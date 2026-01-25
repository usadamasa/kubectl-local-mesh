FROM alpine:3.20

RUN apk add --no-cache \
    bash \
    curl \
    jq

# grpcurl インストール
RUN curl -sSL https://github.com/fullstorydev/grpcurl/releases/download/v1.9.1/grpcurl_1.9.1_linux_x86_64.tar.gz \
    | tar -xz -C /usr/local/bin

WORKDIR /tests

ENTRYPOINT ["bash"]
