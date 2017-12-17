#!/bin/sh

handleConfig()
{
    confType=$1
    etcdIP=$2
    targetDir=$3

    srcFile=""
    dstFile=""
    ##The multus need the yaml, otherwise the cni config.
    case $confType in
        central-node)
            srcFile="/tmp/conf/example-centalip-node.conf"
            dstFile="ovs.conf"
            ;;
        central-cluster)
            srcFile="/tmp/conf/example-centalip-cluster.conf"
            dstFile="ovs.conf"
            ;;
        multus-node)
            srcFile="/tmp/conf/ovs-network-node.yaml"
            dstFile="ovs-network.yaml"
            ;;
        multus-cluster)
            srcFile="/tmp/conf/ovs-network-cluster.yaml"
            dstFile="ovs-network.yaml"
            ;;
        *)
            ;;
    esac
    sed -i "s/127.0.0.1/$etcdIP/g" $srcFile
    cp $srcFile $3/$dstFile
}

confType="centralip-cluster"
etcdIP="127.0.0.1"

while getopts b:t:i:c: option
do
    case "${option}"
        in
        b)  echo "Copy CNI Binary to "${OPTARG}
            cp /tmp/bin/* ${OPTARG}
            ;;
        t)
            confType=`echo ${OPTARG} | cut -d ':' -f1`
            ;;
        i)
            etcdIP=${OPTARG}
            ;;
        c)
            echo "Copty CNI Conf to "${OPTARG}
            handleConfig $confType $etcdIP ${OPTARG}
            ;;
    esac
done

exit

