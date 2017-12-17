FROM alpine:latest

ADD bin/ /tmp/bin
ADD example/ /tmp/conf
ADD example/multus /tmp/conf
ADD Docker/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
