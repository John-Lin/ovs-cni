# ovs-cni

## Testing

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

# Example
ovs-CNI also provides a vagrantfile to help you setup a demo environment to try ovs-CNI.

## Environment
You should install vagrant in your system and make sure everything goes well.

## Setup ovs-CNI
- Change directory to `ovs-CNI` and type `vagrant up` to init a virtual machine.
- Use ssh to connect vagrant VM via `vagrant ssh`.
- Type following commang to build the `ovs-cni` binary and move it to CNI directory.
```
cd ovs-cni
sh build.sh
cp bin/ovs ../cni/ 
```
- We need to provide a CNI config example for `ovs-cni`, and you can use build-in config from example directory. Use following command to copy the config to `~/cni` directory.

```
sudo cp examples/example.conf ~/cni/
```

## Create NS
In this vagrant environment, we don't install docker related services but you can use `namespace(ns)` to test `ovs-cni`.
Type following command to create a namespace named ns1

```
sudo ip netns add ns1
```

## Start CNI
We have setup ovs-CNI environement and some testing namespacese, we can use following command to inform CNI to add a network for namespaces.

```
cd ~/cni
sudo CNI_COMMAND=ADD CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=`pwd` ./ovs <example.conf
```
and the result looks like below
```
{
    "cniVersion": "0.3.1",
        "interfaces": [
        {
            "name": "br0"
        },
        {
            "name": "veth89fa8564"
        },
        {
            "name": "eth2",
            "sandbox": "/var/run/netns/ns2"
        }
        ],
        "ips": [
        {
            "version": "4",
            "address": "10.244.1.12/16",
            "gateway": "10.244.1.1"
        }
        ],
        "routes": [
        {
            "dst": "0.0.0.0/0"
        }
        ],
        "dns": {}
}%
```

Now, we can use some tools to help us check the current network setting, for example.  
You can use `ovs-vsctl show` to show current OVS setting and you it looks like  

```
2a3db108-dc16-44fd-9459-77c6fcb22279
    Bridge "br0"
        fail_mode: standalone
        Port "br0"
            Interface "br0"
                type: internal
        Port "veth656e4f8e"
            Interface "veth656e4f8e"
    ovs_version: "2.5.2"
```

In this setting, the OVS will connect to ns2 via `veth` technology and you can also check
check the namepsace's networking settingm, you can use `sudo ip netns exec ns1 ifconfig` to see its IP config.

```
eth2      Link encap:Ethernet  HWaddr 0a:58:0a:f4:01:0a
          inet addr:10.244.1.10  Bcast:0.0.0.0  Mask:255.255.0.0
          inet6 addr: fe80::bc15:faff:fe6b:b414/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1400  Metric:1
          RX packets:18 errors:0 dropped:0 overruns:0 frame:0
          TX packets:10 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:0
          RX bytes:1476 (1.4 KB)  TX bytes:828 (828.0 B)
`
