# ovs-cni

[![Build Status](https://api.travis-ci.org/John-Lin/ovs-cni.svg?branch=master)](https://travis-ci.org/John-Lin/ovs-cni)
[![codecov](https://codecov.io/gh/John-Lin/ovs-cni/branch/master/graph/badge.svg)](https://codecov.io/gh/John-Lin/ovs-cni)

## Basic Usage

1. Move the binary from `bin/ovs` to `cni/ovs` 
2. Put the example.conf in `cni/example.conf` and run command below

```
$ sudo ip netns add ns1

$ sudo CNI_COMMAND=ADD CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=`pwd` ./ovs <example.conf
```

## Cleanup 
```
$ sudo ip netns del ns1

$ sudo ovs-vsctl del-br br0
```

## Building ovs-cni network plugin

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
and the binary will come out in the `/bin` directory.
