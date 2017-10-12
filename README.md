# ovs-cni

## Testing

1. Move the binary from `bin/ovs` to `cni/ovs` 
2. Put the example.conf in `cni/example.conf` and run command below

```
$ sudo ip netns add ns1

$ sudo CNI_COMMAND=ADD CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=`pwd` ./ovs <example.conf
```

# Cleanup 
```
$ sudo ip netns del ns1

$ sudo ovs-vsctl del-br br0
```