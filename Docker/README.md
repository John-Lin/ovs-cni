## Introduction
In order to use the **kubespray** to automatically deploy the kubenetes cluster with ovs-cni.
We will modify the **kubespray** project and modify the ansible playbook.
In the network config option, we add new config 'ovs' and it will run this dockerfile to copy the pre-build
binary and config into each kubernetes cluster.

## Kubespray
You can find the custom **kubespray** project [here](https://github.com/hwchiu/kubespray).
In that project, just type **vagrant up** and you can use the kubernetes cluster with ovs-cni now.


## How to Build docker image.
Use the **Make docker-build** to build the docker image.

## How to run the docker image
1. First, you should create two directory.
```
mkdir -p /tmp/test/bin
mkdir -p /tmp/test/conf
```
2. run the docker image and mount those two directory
```
sudo docker run -it --rm -v /tmp/test/bin:/opt/cni/bin -v /tmp/test/conf:/etc/cni/net.d/ hwchiu/ovs-cni central-cluster 1.2.3.4
```
3. The entrypoing supports two arguments, IPAM type and etcd IP.
    - IPAM: **central-cluster**/**central-node**
4. After executing the container, you will see two binaries in **/tmp/test/bin** and a ovs.conf in **/tmp/test/conf**
