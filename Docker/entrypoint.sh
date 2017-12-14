#!/bin/sh

cp /cni/bin/ovs /opt/cni/bin/
cp /cni/bin/centralip /opt/cni/bin/

case $1 in
    central-node)
        sed -i "s/127.0.0.1/$2/g" /cni/conf/example-centalip-node.conf
        cp /cni/conf/example-centalip-node.conf /etc/cni/net.d/ovs.conf
        ;;
    central-cluster)
        sed -i "s/127.0.0.1/$2/g" /cni/conf/example-centalip-cluster.conf
        cp /cni/conf/example-centalip-cluster.conf /etc/cni/net.d/ovs.conf
        ;;
    *)
        ;;
esac
