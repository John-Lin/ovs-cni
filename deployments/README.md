# Deployments
Now. We provide three ways to test the ovs-cni plugin.
1. The single host vagrantfile, you can use the docker/namespace to try the ovs-cni plugin.
2. We also provide the vagrantfile for kubernetes cluster, the vagrantfile will create two virtual machines and form a kubernetes cluster and you can use ovs-cni plugin as your kubernetes network.
3. The kubespray script to setup a simple kubernetes cluster with Multus-cni and ovs-cni. The kubespray is a git submodule and you should sync first before you start to use it. After you clone the kubespray project, just type `vagrant up` in that directory and you can use the `vagrant ssh k8s-01` to the k8s master after the installation has completed. 

Go into each directory to learn how to use those vagrantfile.

