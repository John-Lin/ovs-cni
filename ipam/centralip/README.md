## Introduction
The centralip IPAM plugin use the etcd-v3 to handle all ip management.
We provide two mode to decide how to dispatch the ip address.
`Node` mode  and `clster` mode.
In the cluster mode, we even don't provide the gateway address for each node and we hope the SDN controller should handle this, such as ONOS.

## config
The config of centralip like below.
```
   "ipam":{
       "type":"centralip",
       "ipType": "node",
       "network":"10.245.0.0/16",
       "subnetLen": 24,
       "subnetMin": "10.245.5.0",
       "subnetMax": "10.245.50.0",
       "etcdURL": "127.0.0.1:2379"
   }
```
### ipType
We have two backends now, `node` and `cluster`.

In the `node` mode, You need to specify all the following options and it will assign different ip subnet to each node.

In the ohter hand, the `cluster` mode, you only set the `network` and `etcdURL` options.
The `cluster` will assign the IP address to all nodes in the same subnet.
In this mode, we won't provide the gateway address for you, so don't set the `IsDefaultGatway` option in your CNI configuration.
You can see two example configs in the `../../examples`

### network
This field indicate the whole network subnet you want to use.

### subnetLen
This length and `network` will decide the CIDR for each node.
For example, the centralip will dispatch the subnet `10.245.1.0/24`, `10.245.2.0/24`.... `10.245.255.0/24`

### subnetMin/subnetMax
Those two fields is used to indicate the range of subnets you want to dispatch for each node.

### etcdURL
The ip address of etcd-v3 server.
If you want to connect to etcd-v3 servier with TLS, you should also indicate the 
location of three files, including `Certificate`, `Key` and `TrustedCA`.
And you should use the following key to describe your location to CNI.
```bash
       "etcdURL": "https://127.0.0.1:2379",
       "etcdCertFile": "/etc/ovs/certs/cert.crt",
       "etcdKeyFile": "/etc/ovs/certs/key.pem",
       "etcdTrustedCAFileFile": "/etc/ovs/certs/ca_cert.crt"
```
