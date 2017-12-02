# Kubernetes in vagrant

This document intends to give a instruction about how to create a Kubernetes cluster with ovs-cni in vagrant.
The Kubernetes cluster is created by kubeadm but still has some part should be handled it by manually. 

## Table of Contents

- [Kubernetes setup in vagrant](#kubernetes-setup-in-vagrant)
- [Use ovs-cni as a default network interface](#use-ovs-cni-as-a-default-network-interface)
- [Extend the multiple network interfaces with OVS CNI](#extend-the-multiple-network-interfaces-with-ovs-cni)

## Kubernetes setup in vagrant

### Before vagrant up

Check the default network interface on your machine. Because we used the bridged networking in Virtualbox which uses a device driver on your **host** system.

In my case the default network interface is **en0**. Mdify to your default network interface before go to next step.

```sh
host1.vm.network "public_network", bridge: "en0: Wi-Fi (AirPort)"
...
...
...
host2.vm.network "public_network", bridge: "en0: Wi-Fi (AirPort)"
```
[Default Network Interface](https://www.vagrantup.com/docs/networking/public_network.html#default-network-interface)

### Vagrant up

Use `vagrant up` to create a two nodes Kubernetes cluster. The detail bootstrap script is in the `Vagrantfile` and `k8s-bootstrap.sh`. 
It will

- install Go
- install Docker
- install kubelet kubeadm kubectl
- download CNI and CNI plugins binaries with version 0.6.0 and replace the old one.
- download OVS CNI plugins source
- build the ovs-cni binary and copy to `/opt/cni/bin/`

in each virtual machine.

### Vagrant ssh

Use `vagrant ssh HOSTNAME` to ssh into virtaul machine. **HOSTNAME** could be define in the Vagrantfile.

### Before kubeadm init

Check the IP on the each virtual machine. Choose a network interface that has just attach to Virtualbox's bridge.

In my case the `enp0s8` is the one that attached to Virtualbox's bridge. Use `192.168.0.107` as our Kubernetes master IP address.

On host1
```
$ vagrant ssh host1

$ ifconfig
enp0s3    Link encap:Ethernet  HWaddr 02:a4:a0:f1:db:6e
          inet addr:10.0.2.15  Bcast:10.0.2.255  Mask:255.255.255.0
          inet6 addr: fe80::a4:a0ff:fef1:db6e/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:377418 errors:0 dropped:0 overruns:0 frame:0
          TX packets:151364 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:471642757 (471.6 MB)  TX bytes:9624180 (9.6 MB)

enp0s8    Link encap:Ethernet  HWaddr 08:00:27:27:ba:02
          inet addr:192.168.0.107  Bcast:192.168.0.255  Mask:255.255.255.0
          inet6 addr: fe80::a00:27ff:fe27:ba02/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:92934 errors:0 dropped:0 overruns:0 frame:0
          TX packets:89463 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:9514803 (9.5 MB)  TX bytes:62806152 (62.8 MB)
```

And use `192.168.0.108` as our Kubernetes node/minion IP address.

On host2
```
$ vagrant ssh host2

$ ifconfig
enp0s3    Link encap:Ethernet  HWaddr 02:a4:a0:f1:db:6e
          inet addr:10.0.2.15  Bcast:10.0.2.255  Mask:255.255.255.0
          inet6 addr: fe80::a4:a0ff:fef1:db6e/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:268508 errors:0 dropped:0 overruns:0 frame:0
          TX packets:107769 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:320038637 (320.0 MB)  TX bytes:6793921 (6.7 MB)

enp0s8    Link encap:Ethernet  HWaddr 08:00:27:f2:36:fb
          inet addr:192.168.0.108  Bcast:192.168.0.255  Mask:255.255.255.0
          inet6 addr: fe80::a00:27ff:fef2:36fb/64 Scope:Link
          UP BROADCAST RUNNING MULTICAST  MTU:1500  Metric:1
          RX packets:90239 errors:0 dropped:0 overruns:0 frame:0
          TX packets:94242 errors:0 dropped:0 overruns:0 carrier:0
          collisions:0 txqueuelen:1000
          RX bytes:45148814 (45.1 MB)  TX bytes:9918777 (9.9 MB)
```

next, edit and add the `--node-ip` option in the configuration of `10-kubeadm.conf`

append `Environment="KUBELET_EXTRA_ARGS=--node-ip=192.168.0.108"` in the file on **BOTH virtual machines**

The `192.168.0.108` is our node/minion IP address.

```
sudo vim /etc/systemd/system/kubelet.service.d/10-kubeadm.conf

[Service]
...
...
..
Environment="KUBELET_EXTRA_ARGS=--node-ip=192.168.0.108"
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_SYSTEM_PODS_ARGS $KUBELET_NETWORK_ARGS $KUBELET_DNS_ARGS $KUBELET_AUTHZ_ARGS $KUBELET_CADVISOR_ARGS $KUBELET_CERTIFICATE_ARGS $KUBELET_EXTRA_ARGS
```

### Kubeadm init

Run the `kubeadm init` to initializing master. In the host1 (Master) run `kubeadm init` with option `--apiserver-advertise-address` this will specify the correct interface to advertise the master on as the interface with the default gateway.

```
sudo kubeadm init --apiserver-advertise-address=192.168.0.107
...
...
Your Kubernetes master has initialized successfully!

To start using your cluster, you need to run (as a regular user):

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  http://kubernetes.io/docs/admin/addons/

You can now join any number of machines by running the following on each node
as root:

  kubeadm join --token <token> <master-ip>:<master-port> --discovery-token-ca-cert-hash sha256:<hash>
```

To join the machines use `kubeadm join` on host2.

### After join number of machines by running kubeadm join

Restart kubelet on **BOTH virtual machines** to activate the `KUBELET_EXTRA_ARGS` has configured before:

```
$ sudo systemctl daemon-reload
$ sudo systemctl restart kubelet
```

## Use ovs-cni as a default network interface

### Configuring ovs-cni network plugin

Modify the configuration to meet your requirements. Check the `example.conf` and copy into the default `--cni-conf-dir` path which is in `/etc/cni/net.d`.

For the host1, given the following network configuration:

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
# tee /etc/cni/net.d/ovs.conf <<-'EOF'
{
   "name":"mynet",
   "cniVersion":"0.3.1",
   "type":"ovs",
   "ovsBridge":"br0",
   "vtepIPs":[
      "192.168.0.108"
   ],
   "isDefaultGateway": true,
   "ipMasq": true,
   "ipam":{
      "type":"host-local",
      "subnet":"10.244.0.0/16",
      "rangeStart":"10.244.1.10",
      "rangeEnd":"10.244.1.150",
      "routes":[
         {
            "dst":"0.0.0.0/0"
         }
      ],
      "gateway":"10.244.1.1"
   }
}
EOF
```

For the host2, given the following network configuration:

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
# tee /etc/cni/net.d/ovs.conf <<-'EOF'
{
   "name":"mynet",
   "cniVersion":"0.3.1",
   "type":"ovs",
   "ovsBridge":"br0",
   "vtepIPs":[
      "192.168.0.107"
   ],
   "isDefaultGateway": true,
   "ipMasq": true,
   "ipam":{
      "type":"host-local",
      "subnet":"10.244.0.0/16",
      "rangeStart":"10.244.2.10",
      "rangeEnd":"10.244.2.150",
      "routes":[
         {
            "dst":"0.0.0.0/0"
         }
      ],
      "gateway":"10.244.2.1"
   }
}
EOF
```

**Note**: the `vtepIPs`, `rangeStart`, `rangeEnd` and `gateway` could be different on each host.

### Master Isolation

By default, your cluster will not schedule pods on the master for security reasons. If you want to be able to schedule pods on the master, e.g. for a single-machine Kubernetes cluster for development, run:

```
$ kubectl taint nodes --all node-role.kubernetes.io/master-
```

### Apply the deployment on Kubernetes cluster

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
$ kubectl apply -f ./kubernetes/deployments/busybox.yaml
$ kubectl get pod -o wide
NAME                                  READY     STATUS    RESTARTS   AGE       IP            NODE
busybox-deployment-6b8c55d957-6wjcl   1/1       Running   11         1d        10.244.2.11   host2
busybox-deployment-6b8c55d957-pxm6c   1/1       Running   11         1d        10.244.1.12   host1
```

## Extend the multiple network interfaces with OVS CNI

### Build and install Multus plugin

This sould be do on ALL kubernetes nodes

### Building multus
```shell
git cloen https://github.com/Intel-Corp/multus-cni.git
cd multus-cni
./build
```

### Installation
Copy the binary from `/bin/multus` to `/opt/cni/bin/` and make sure the `ovs` binary is inside the directory

Create Multus CNI configuration file `/etc/cni/net.d/multus-cni.conf` with below content in minions. Use only the absolute path to point to the kubeconfig file (it may change depending upon your cluster env) and make sure all CNI binary files are in `/opt/cni/bin` dir

```
{
    "name": "minion-cni-network",
    "type": "multus",
    "kubeconfig": "/home/ubuntu/.kube/config",
    "delegates": [{
        "type": "ovs",
        "masterplugin": true
    }]
}
```

You might need to copy kubeconfig file from the kubernetes master node `/home/ubuntu/.kube/config` to all minion nodes. This could allow Multus works.

### Create a Custom Resource Definition (CRD) based Network objects

Create a Third party resource `crdnetwork.yaml` for the network object

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: networks.kubernetes.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: kubernetes.com
  # version name to use for REST API: /apis/<group>/<version>
  version: v1
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: networks
    # singular name to be used as an alias on the CLI and for display
    singular: network
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: Network
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - net
```

kubectl create command for the Custom Resource Definition

```sh
# kubectl create -f ./crdnetwork.yaml
customresourcedefinition "network.kubernetes.com" created
```

(kubectl get command to check the Network CRD creation)
```sh
# kubectl get CustomResourceDefinition
# kubectl get crd
NAME                      KIND
networks.kubernetes.com   CustomResourceDefinition.v1beta1.apiextensions.k8s.io
```

Save the below following YAML to ovs-network.yaml
```yaml
apiVersion: "kubernetes.com/v1"
kind: Network
metadata:
  name: ovs-net
plugin: ovs
args: '[
        {
        "name": "myovsnet",
        "type": "ovs",
        "ovsBridge":"br0",
        "isDefaultGateway": true,
        "ipMasq": true,
        "ipam":{
            "type":"centralip",
            "network":"10.245.0.0/16",
            "subnetLen": 24,
            "subnetMin": "10.245.5.0",
            "subnetMax": "10.245.50.0",
            "etcdURL": "192.168.0.107:2379"
        }
        }
]'
```
With ipam type `centralip` should setup a ETCD server. By default it should be set to kubernetes master server IP.

Create the ovs network object

```shell
# kubectl create -f ovs-network.yaml
network "ovs-net" created
```

Check the network object

```shell
# kubectl get network
# kubectl get net
```


Next, modify the etcd manifests to allow etcd server running on public (from 127.0.0.1 to 0.0.0.0) in kuberneter master node (This could cause security issue)

```shell
sudo vim /etc/kubernetes/manifests/etcd.yaml
```

```yaml
...
spec:
  containers:
  - command:
    - etcd
    - --listen-client-urls=http://0.0.0.0:2379
    - --advertise-client-urls=http://0.0.0.0:2379
    - --data-dir=/var/lib/etcd
...
...
```

Restart kubelet on master node

```
$ sudo systemctl daemon-reload
$ sudo systemctl restart kubelet
```


### Configuring Pod to use the CRD Network objects


Save the below following YAML to pod-multi-network.yaml. In this case `flannel-conf` network object act as the primary network.

```yaml
# cat pod-multi-network.yaml 
apiVersion: v1
kind: Pod
metadata:
  name: multus-multi-net-poc
  annotations:
    networks: '[  
        { "name": "ovs-net" },
        { "name": "flannel-conf" }
    ]'
spec:  # specification of the pod's contents
  containers:
  - name: multus-multi-net-poc
    image: "busybox"
    command: ["top"]
    stdin: true
    tty: true
```

For setting up `flannel-conf` please see

https://github.com/coreos/flannel/blob/4c057be1f97a38960436834f144b0bd824d7f76e/README.md#multi-network-mode-experimental

https://github.com/coreos/flannel/blob/master/Documentation/running.md

Create Multiple network based pod from the master node

```shell
# kubectl create -f ./pod-multi-network.yaml
pod "multus-multi-net-poc" created
```


### References 

- https://github.com/Intel-Corp/multus-cni
- https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-resources
