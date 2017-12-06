# ovs-cni

[![Build Status](https://api.travis-ci.org/John-Lin/ovs-cni.svg?branch=master)](https://travis-ci.org/John-Lin/ovs-cni)
[![codecov](https://codecov.io/gh/John-Lin/ovs-cni/branch/master/graph/badge.svg)](https://codecov.io/gh/John-Lin/ovs-cni)

## Introduction

OVS-CNI is a totally Open vSwitch CNI plugin, it create the Open vSwitch and use veth to connect the OpenvSwitch and container.

ovs-cni supports the following options in its config and you can see the example config `example/example.conf` to learn how to use it.

1. controller IP

```
controller:10.245.1.5:6653
```

2. target vxlan IP

```
vtepIPs: [
    "10.245.2.2",
    "10.245.2.3"
]
```

3. bridge name

```
ovsBridge: "br0"
```

4. Act as a gateway for pods.

```
isDefaultGateway: true
```

5. Support SNAT for pods traffic.

```
ipMasq: true
```

6. IPAM support

ovs-cni support basic IPAM type such as host-local, you can see `example/example.conf` to see how config it.
Besides, ovs-cni provide a new IPAM plugin central-ip, which use the `ETCD` to perform centralized IP assignment/management and you can go to `ipam/centralip` directory to see more usage about it.

## Usage

If you are familiar with CNI plugin and know how to use it. you can refer to the following instruction to use it.
Otherwise you can go to `deployment` directory to learn how to use it.

### Building ovs-cni

Because ovs-cni used the package management tool called `govendor`, we have to install the govendor first.

```
$ go get -u github.com/kardianos/govendor
```

We use `govendor` to download all dependencies

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
$ govendor sync
```

build the ovs-cni binary.

```
$ ./build.sh
```
and the binary will come out in the `/bin` directory and you can find `ovs` and `centralip`.
The `ovs` is the main CNI plugin and the `centralip` is the CNI plugin for different IPAM usage.
If you want to use `ETCD` to centralizaed manage the IP address, you should also copy the `centralip` binary to the CNI directory and modify the config to use it.

```
$ sudo ip netns add ns1

$ sudo CNI_COMMAND=ADD CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=`pwd` ./ovs <example.conf
```

## Use the ovs-cni

Before running the ovs-cni, make sure your system meet the following requirement.

1. install openvswitch and run it.

2. download the basic cni binary (local-host) and put they in `/opt/cni/bin`

Now, we use `ip netns` to create a basic network namespace as our example.

```
$ sudo ip netns add ns1
```

change directory to bin and type following command to start the ovs-cni.

```
$ sudo CNI_COMMAND=DEL CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=/opt/cni/bin ./ovs < ../example/example.conf
$ sudo ip netns exec ns1 ifconfig
```

## Cleanup

```
$ sudo ip netns del ns1

$ sudo ovs-vsctl del-br br0
```
