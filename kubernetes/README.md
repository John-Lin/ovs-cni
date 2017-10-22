# Kubernetes in vagrant

This document intends to give a instruction about how to create a Kubernetes cluster with ovs-cni in vagrant.
The Kubernetes cluster is created by kubeadm but still has some part we should do it by manually. 

## Before vagrant up

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

## Vagrant up 

Use `vagrant up` to create a two nodes Kubernetes cluster. The detail bootstrap script is in the `Vagrantfile` and `k8s-bootstrap.sh`. 
It will

- install Go
- install Docker
- install kubelet kubeadm kubectl
- download CNI and CNI plugins binaries with version 0.6.0 and replace the old one.
- download OVS CNI plugins source

in each virtual machine.

## Vagrant ssh

Use `vagrant ssh HOSTNAME` and `vagrant ssh HOSTNAME` to ssh into virtaul machine. **HOSTNAME** could be define in the Vagrantfile.

## Before kubeadm init

We should check the IP on the each virtual machine. Choose a network interface that we've just attach to Virtualbox's bridge.

In my case the `enp0s8` is the one that attached to Virtualbox's bridge. We will use `192.168.0.107` as our Kubernetes master IP address.

On host2
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

And we will use `192.168.0.108` as our Kubernetes node/minion IP address.

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

add `Environment="KUBELET_EXTRA_ARGS=--node-ip=192.168.0.108"` in the file on **BOTH virtual machines**

The `192.168.0.108` is our node/minion IP address. we've previously mentioned.

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

## Kubeadm init

Now we can run the `kubeadm init` to initializing your master. In the host1 (Master) run `kubeadm init` with option `--apiserver-advertise-address` this will specify the correct interface to advertise the master on as the interface with the default gateway.

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

## After join number of machines by running kubeadm join

We have to restart kubelet on **BOTH virtual machines** to activate the `KUBELET_EXTRA_ARGS` we've configured before:

```
$ sudo systemctl daemon-reload
$ sudo systemctl restart kubelet
```

## Building ovs-cni network plugin

Because ovs-cni used the package management tool called `govendor`, we have to install the govendor first.

```
$ go get -u github.com/kardianos/govendor
```

Then, we use `govendor` to download all dependencies

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
$ govendor sync
```

build the ovs-cni binary.

```
$ ./build
```
the binary will come out in the `/bin` directory.

## Installing ovs-cni network plugin

To install the network plugin we have to copy the binary to the default `--cni-bin-dir` path which is in `/opt/cni/bin`.

This should be done on **BOTH virtual machines**

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
$ sudo cp ./bin/ovs /opt/cni/bin
```

Next, modify the configuration to meet your requirements. Check the `example.conf` and copy into the default `--cni-conf-dir` path which is in `/etc/cni/net.d`.

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

## Apply the deployment on Kubernetes cluster

```
$ cd ~/go/src/github.com/John-Lin/ovs-cni
$ kubectl apply -f ./kubernetes/deployments/busybox.yaml
$ kubectl get pod -o wide
NAME                                  READY     STATUS    RESTARTS   AGE       IP            NODE
busybox-deployment-6b8c55d957-6wjcl   1/1       Running   11         1d        10.244.2.11   host2
busybox-deployment-6b8c55d957-pxm6c   1/1       Running   11         1d        10.244.1.12   host1
```