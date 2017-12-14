FROM alpine:latest

ADD bin/ovs /cni/bin/ovs
ADD bin/centralip /cni/bin/centralip
COPY Docker/entrypoint.sh /
COPY example/ /cni/conf/
ENTRYPOINT ["/entrypoint.sh"]
