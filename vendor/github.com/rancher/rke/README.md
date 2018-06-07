# rke

Rancher Kubernetes Engine, an extremely simple, lightning fast Kubernetes installer that works everywhere.

## Download

Please check the [releases](https://github.com/rancher/rke/releases/) page.

## Requirements

- Docker versions `1.11.2` up to `1.13.1` and `17.03.x` are validated for Kubernetes versions 1.8, 1.9 and 1.10
- OpenSSH 7.0+ must be installed on each node for stream local forwarding to work.
- The SSH user used for node access must be a member of the `docker` group:

```bash
usermod -aG docker <user_name>
```

- Ports 6443, 2379, and 2380 should be opened between cluster nodes.
- Swap disabled on worker nodes.

## Getting Started

Starting out with RKE? Check out this [blog post](http://rancher.com/an-introduction-to-rke/) or the [Quick Start Guide](https://github.com/rancher/rke/wiki/Quick-Start-Guide)

## Using RKE

Standing up a Kubernetes is as simple as creating a `cluster.yml` configuration file and running the command:

```bash
./rke up --config cluster.yml
```

### Full `cluster.yml` example

You can view full sample of cluster.yml [here](https://github.com/rancher/rke/blob/master/cluster.yml).

### Minimal `cluster.yml` example

```yaml
# default k8s version: v1.8.10-rancher1-1
# default network plugin: canal
nodes:
  - address: 1.2.3.4
    user: ubuntu
    role: [controlplane,worker,etcd]

```

## Kubernetes Version

The current default kubernetes version used by RKE is `v1.10.1-rancher1`.

There are two ways to select a kubernetes version:

- Using the kubernetes image defined in [System Images](#rke-system-images)
- Using the configuration option `kubernetes_version`

In case both are defined, the system images configuration will take precedence over `kubernetes_version`. Since the `kubernetes_version` options was added mainly to be used by Rancher v2.0, it has a limited number of supported tags that can be found [here](https://github.com/rancher/types/blob/master/apis/management.cattle.io/v3/k8s_defaults.go#L14).

If a version is defined in `kubernetes_version` and is not found in this map, the default is used.

## Network Plugins

RKE supports the following network plugins that are deployed as addons:

- Flannel
- Calico
- Canal
- Weave

To use specific network plugin configure `cluster.yml` to include:

```yaml
network:
  plugin: flannel
```

### Network Options

There are extra options that can be specified for each network plugin:

#### Flannel

- **flannel_image**: Flannel daemon Docker image
- **flannel_cni_image**: Flannel CNI binary installer Docker image
- **flannel_iface**: Interface to use for inter-host communication

#### Calico

- **calico_node_image**: Calico Daemon Docker image
- **calico_cni_image**: Calico CNI binary installer Docker image
- **calico_controllers_image**: Calico Controller Docker image
- **calicoctl_image**: Calicoctl tool Docker image
- **calico_cloud_provider**: Cloud provider where Calico will operate, currently supported values are: `aws`, `gce`

#### Canal

- **canal_node_image**: Canal Node Docker image
- **canal_cni_image**: Canal CNI binary installer Docker image
- **canal_flannel_image**: Canal Flannel Docker image

#### Weave

- **weave_node_image**: Weave Node Docker image
- **weave_cni_image**: Weave CNI binary installer Docker image

### RKE System Images

Prior to version `0.1.6`, RKE used the following list of images for deployment and cluster configuration:
```
system_images:
  etcd: rancher/etcd:v3.0.17
  kubernetes: rancher/k8s:v1.8.9-rancher1-1
  alpine: alpine:latest
  nginx_proxy: rancher/rke-nginx-proxy:v0.1.1
  cert_downloader: rancher/rke-cert-deployer:v0.1.1
  kubernetes_services_sidecar: rancher/rke-service-sidekick:v0.1.0
  kubedns: rancher/k8s-dns-kube-dns-amd64:1.14.5
  dnsmasq: rancher/k8s-dns-dnsmasq-nanny-amd64:1.14.5
  kubedns_sidecar: rancher/k8s-dns-sidecar-amd64:1.14.5
  kubedns_autoscaler: rancher/cluster-proportional-autoscaler-amd64:1.0.0
  flannel: rancher/coreos-flannel:v0.9.1
  flannel_cni: rancher/coreos-flannel-cni:v0.2.0
```
As of version `0.1.6`, we consolidated several of those images into a single image to simplify and speed the deployment process.

The following images are no longer required, and can be replaced by `rancher/rke-tools:v0.1.4`:
- alpine:latest
- rancher/rke-nginx-proxy:v0.1.1
- rancher/rke-cert-deployer:v0.1.1
- rancher/rke-service-sidekick:v0.1.0

## Addons

RKE supports pluggable addons. Addons are used to deploy several cluster components including:
- Network plugin
- KubeDNS
- Ingress controller

In addition, a user can specify the addon yaml in the cluster.yml file, and when running

```yaml
rke up --config cluster.yml
```

RKE will deploy the addons yaml after the cluster starts, RKE first uploads this yaml file as a configmap in kubernetes cluster and then run a kubernetes job that mounts this config map and deploy the addons.

> Note that RKE doesn't support yet removal or update of the addons, so once they are deployed the first time you can't change them using rke

To start using addons use `addons:` option in the `cluster.yml` file for example:

```yaml
addons: |-
    ---
    apiVersion: v1
    kind: Pod
    metadata:
      name: my-nginx
      namespace: default
    spec:
      containers:
      - name: my-nginx
        image: nginx
        ports:
        - containerPort: 80
```

Note that we are using `|-` because the addons option is a multi line string option, where you can specify multiple yaml files and separate them with `---`

For `addons_include:` you may pass either http/https urls or file paths, for example:
```yaml
addons_include:
    - https://raw.githubusercontent.com/rook/rook/master/cluster/examples/kubernetes/rook-operator.yaml
    - https://raw.githubusercontent.com/rook/rook/master/cluster/examples/kubernetes/rook-cluster.yaml
    - /opt/manifests/example.yaml
    - ./nginx.yaml
```

#### Addon deployment jobs

RKE uses kubernetes Jobs to deploy addons. In some cases, addons deployment takes longer than expected. Starting with version `0.1.7-rc1`, RKE provides an option to controle the job check timeout in seconds:
```yaml
addon_job_timeout: 30
```

#### Critical and noncritical addons
As of version `0.1.7-rc1`, addons are split into two categories: critical and noncritical.

Critical addons will cause RKE to error out if they fail to deploy for any reason. While noncritical addons will just log a warning and continue with the deployment. Currently only the network plugin is considered critical.
## High Availability

RKE is HA ready, you can specify more than one controlplane host in the `cluster.yml` file, and rke will deploy master components on all of them, the kubelets are configured to connect to `127.0.0.1:6443` by default which is the address of `nginx-proxy` service that proxy requests to all master nodes.

to start an HA cluster, just specify more than one host with role `controlplane`, and start the cluster normally.

## Adding/Removing Nodes

RKE supports adding/removing nodes for worker and controlplane hosts, in order to add additional nodes you will only need to update the `cluster.yml` file with additional nodes and run `rke up` with the same file.

To remove nodes just remove them from the hosts list in the cluster configuration file `cluster.yml`, and re run `rke up` command.

## Cluster Remove

RKE supports `rke remove` command, the command does the following:

- Connect to each host and remove the kubernetes services deployed on it.
- Clean each host from the directories left by the services:
  - /etc/kubernetes/ssl
  - /var/lib/etcd
  - /etc/cni
  - /opt/cni
  - /var/run/calico

Note that this command is irreversible and will destroy the kubernetes cluster entirely.

## Cluster Upgrade

RKE supports kubernetes cluster upgrade through changing the image version of services, in order to do that change the image option for each services, for example:

```yaml
image: rancher/hyperkube:v1.9.7
```

TO

```yaml
image: rancher/hyperkube:v1.10.1
```

And then run:

```bash
rke up --config cluster.yml
```

RKE will first look for the local `kube_config_cluster.yml` and then tries to upgrade each service to the latest image.

> Note that rollback isn't supported in RKE and may lead to unxpected results

## Service Upgrade

Service can also be upgraded by changing any of the services arguments or extra args and run `rke up` again with the updated configuration file.

> Please note that changing the following arguments: `service_cluster_ip_range` or `cluster_cidr` will result in a broken cluster, because currently the network pods will not be automatically upgraded.

## RKE Config

RKE supports command `rke config` which generates a cluster config template for the user, to start using this command just write:

```bash
rke config --name mycluster.yml
```

RKE will ask some questions around the cluster file like number of the hosts, ips, ssh users, etc, `--empty` option will generate an empty cluster.yml file, also if you just want to print on the screen and not save it in a file you can use `--print`.

## Ingress Controller

RKE will deploy Nginx controller by default, user can disable this by specifying `none` to ingress `provider` option in the cluster configuration, user also can specify list of options for nginx config map listed in this [doc](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/configmap.md), and command line extra_args listed in this [doc](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/cli-arguments.md), for example:
```
ingress:
  provider: nginx
  options:
    map-hash-bucket-size: "128"
    ssl-protocols: SSLv2
  extra_args:
    enable-ssl-passthrough: ""
```
By default, RKE will deploy ingress controller on all schedulable nodes (controlplane and workers), to specify only certain nodes for ingress controller to be deployed, user has to specify `node_selector` for the ingress and the right label on the node, for example:
```
nodes:
  - address: 1.1.1.1
    role: [controlplane,worker,etcd]
    user: root
    labels:
      app: ingress

ingress:
  provider: nginx
  node_selector:
    app: ingress
```

RKE will deploy Nginx Ingress controller as a DaemonSet with `hostnetwork: true`, so ports `80`, and `443` will be opened on each node where the controller is deployed.

## Extra Args and Binds

RKE supports additional service arguments.

```yaml
services:
  # ...
  kube-controller:
    extra_args:
      cluster-name: "mycluster"
```
This will add/append `--cluster-name=mycluster` to the container list of arguments.

As of `v0.1.3-rc2` using `extra_args` will add new arguments and **override** existing defaults. For example, if you need to modify the default admission controllers list, you need to change the default list and add apply it using `extra_args`.

RKE also supports additional volume binds:

```yaml
services:
  # ...
  kubelet:
    extra_binds:
      - "/host/dev:/dev"
      - "/usr/libexec/kubernetes/kubelet-plugins:/usr/libexec/kubernetes/kubelet-plugins:z"
```

## Authentication

RKE Supports x509 authentication strategy. You can additionally define a list of SANs (Subject Alternative Names) to add to the Kubernetes API Server PKI certificates. This allows you to connect to your Kubernetes cluster API Server through a load balancer, for example, rather than a single node.

```yaml
authentication:
  strategy: x509
  sans:
  - "10.18.160.10"
  - "my-loadbalancer-1234567890.us-west-2.elb.amazonaws.com"
```

## External etcd

RKE supports using external etcd instead of deploying etcd servers, to enable external etcd the following parameters should be populated:

```
services:
  etcd:
    path: /etcdcluster
    external_urls:
      - https://etcd-example.com:2379
    ca_cert: |-
      -----BEGIN CERTIFICATE-----
      xxxxxxxxxx
      -----END CERTIFICATE-----
    cert: |-
      -----BEGIN CERTIFICATE-----
      xxxxxxxxxx
      -----END CERTIFICATE-----
    key: |-
      -----BEGIN PRIVATE KEY-----
      xxxxxxxxxx
      -----END PRIVATE KEY-----
```

Note that RKE only supports connecting to TLS enabled etcd setup, user can enable multiple endpoints in the `external_urls` field. RKE will not accept having external urls and nodes with `etcd` role at the same time, user should only specify either etcd role for servers or external etcd but not both.

## Cloud Providers

Starting from v0.1.3 rke supports cloud providers.

### AWS Cloud Provider

To enable AWS cloud provider, you can set the following in the cluster configuration file:
```
cloud_provider:
  name: aws
```

AWS cloud provider has to be enabled on ec2 instances with the right IAM role.

### Azure Cloud provider

Azure cloud provider can be enabled by passing `azure` as the cloud provider name and set of options to the configuration file:
```
cloud_provider:
  name: azure
  cloud_config:
    aadClientId: xxxxxxxxxxxx
    aadClientSecret: xxxxxxxxxxx
    location: westus
    resourceGroup: rke-rg
    subnetName: rke-subnet
    subscriptionId: xxxxxxxxxxx
    vnetName: rke-vnet
    tenantId: xxxxxxxxxx
    securityGroupName: rke-nsg
```

You also have to make sure that the Azure node name must match the kubernetes node name, you can do that by changing the value of hostname_override in the config file:
```
nodes:
  - address: x.x.x.x
    hostname_override: azure-rke1
    user: ubuntu
    role:
    - controlplane
    - etcd
    - worker
```

## Deploying Rancher 2.x using rke

Using RKE's pluggable user addons, it's possible to deploy Rancher 2.x server in HA with a single command. Detailed instructions can be found [here](https://rancher.com/docs/rancher/v2.x/en/installation/ha-server-install/).

## Operating Systems Notes

### Atomic OS

- Container volumes may have some issues in Atomic OS due to SELinux, most of volumes are mounted in rke with option `z`, however user still need to run the following commands before running rke:
```
# mkdir /opt/cni /etc/cni
# chcon -Rt svirt_sandbox_file_t /etc/cni
# chcon -Rt svirt_sandbox_file_t /opt/cni
```
- OpenSSH 6.4 shipped by default on Atomic CentOS which doesn't support SSH tunneling and therefore breaks rke, upgrading OpenSSH to the latest version supported by Atomic host will solve this problem:
```
# atomic host upgrade
```
- Atomic host doesn't come with docker group by default, you can change ownership of docker.sock to enable specific user to run rke:
```
# chown <user> /var/run/docker.sock
```

## Etcd Snapshots

You can configure a Rancher Kubernetes Engine (RKE) cluster to automatically take snapshots of etcd. In a disaster scenario, you can restore these snapshots, which are stored on other cluster nodes.

### One-Time Snapshots

RKE introduce a new command that can take a snapshot of a running etcd node in rke cluster, the snapshot will be automatically saved in `/opt/rke/etcd-snapshots`, the commands works as following:
```
./rke etcd snapshot-save --config cluster.yml     

WARN[0000] Name of the snapshot is not specified using [rke_etcd_snapshot_2018-05-17T23:32:08+02:00]
INFO[0000] Starting saving snapshot on etcd hosts       
INFO[0000] [dialer] Setup tunnel for host [x.x.x.x]
INFO[0001] [dialer] Setup tunnel for host [y.y.y.y]
INFO[0002] [dialer] Setup tunnel for host [z.z.z.z]
INFO[0003] [etcd] Saving snapshot [rke_etcd_snapshot_2018-05-17T23:32:08+02:00] on host [x.x.x.x]
INFO[0004] [etcd] Successfully started [etcd-snapshot-once] container on host [x.x.x.x]
INFO[0004] [etcd] Saving snapshot [rke_etcd_snapshot_2018-05-17T23:32:08+02:00] on host [y.y.y.y]
INFO[0005] [etcd] Successfully started [etcd-snapshot-once] container on host [y.y.y.y]
INFO[0005] [etcd] Saving snapshot [rke_etcd_snapshot_2018-05-17T23:32:08+02:00] on host [z.z.z.z]
INFO[0006] [etcd] Successfully started [etcd-snapshot-once] container on host [z.z.z.z]
INFO[0006] Finished saving snapshot [rke_etcd_snapshot_2018-05-17T23:32:08+02:00] on all etcd hosts
```

The command will save a snapshot of etcd from each etcd node in the cluster config file and will save it in `/opt/rke/etcd-snapshots`. This command also creates a container for taking the snapshot. When the process completes, the container is automatically removed.

### Etcd Recurring Snapshots

To schedule a recurring automatic etcd snapshot save, enable the `etcd-snapshot` service. `etcd-snapshot` runs in a service container alongside the `etcd` container. `etcd-snapshot` automatically takes a snapshot of etcd and stores them to its local disk in `/opt/rke/etcd-snapshots`.

To enable `etcd-snapshot` in RKE CLI, configure the following three variables:

```
services:
  etcd:
    snapshot: true
    creation: 5m0s
    retention: 24h
```

- `snapshot`: Enables/disables etcd snapshot recurring service in the RKE cluster.

	Default value: `false`.
- `creation`: Time period in which `etcd-sanpshot` take snapshots.

	Default value: `5m0s`

- `retention`: Time period before before an etcd snapshot expires. Expired snapshots are purged.

	Default value: `24h`

After RKE runs, view the `etcd-snapshot` logs to confirm backups are being created automatically:
```
# docker logs etcd-snapshot

time="2018-05-04T18:39:16Z" level=info msg="Initializing Rolling Backups" creation=1m0s retention=24h0m0s
time="2018-05-04T18:40:16Z" level=info msg="Created backup" name="2018-05-04T18:40:16Z_etcd" runtime=108.332814ms
time="2018-05-04T18:41:16Z" level=info msg="Created backup" name="2018-05-04T18:41:16Z_etcd" runtime=92.880112ms
time="2018-05-04T18:42:16Z" level=info msg="Created backup" name="2018-05-04T18:42:16Z_etcd" runtime=83.67642ms
time="2018-05-04T18:43:16Z" level=info msg="Created backup" name="2018-05-04T18:43:16Z_etcd" runtime=86.298499ms
```
Backups are saved to the following directory: `/opt/rke/etcd-snapshots/`. Backups are created on each node that runs etcd.


### Etcd Disaster recovery

`etcd snapshot-restore` is used for etcd Disaster recovery, it reverts to any snapshot stored in `/opt/rke/etcd-snapshots` that you explicitly define. When you run `etcd snapshot-restore`, RKE removes the old etcd container if it still exists. To restore operations, RKE creates a new etcd cluster using the snapshot you choose.

>**Warning:** Restoring an etcd snapshot deletes your current etcd cluster and replaces it with a new one. Before you run the `etcd snapshot-restore` command, backup any important data in your current cluster.

```
./rke etcd snapshot-restore --name snapshot --config cluster.yml
INFO[0000] Starting restore on etcd hosts
INFO[0000] [dialer] Setup tunnel for host [x.x.x.x]
INFO[0002] [dialer] Setup tunnel for host [y.y.y.y]
INFO[0005] [dialer] Setup tunnel for host [z.z.z.z]
INFO[0007] [hosts] Cleaning up host [x.x.x.x]
INFO[0007] [hosts] Running cleaner container on host [x.x.x.x]
INFO[0008] [kube-cleaner] Successfully started [kube-cleaner] container on host [x.x.x.x]
INFO[0008] [hosts] Removing cleaner container on host [x.x.x.x]
INFO[0008] [hosts] Successfully cleaned up host [x.x.x.x]
INFO[0009] [hosts] Cleaning up host [y.y.y.y]
INFO[0009] [hosts] Running cleaner container on host [y.y.y.y]
INFO[0010] [kube-cleaner] Successfully started [kube-cleaner] container on host [y.y.y.y]
INFO[0010] [hosts] Removing cleaner container on host [y.y.y.y]
INFO[0010] [hosts] Successfully cleaned up host [y.y.y.y]
INFO[0011] [hosts] Cleaning up host [z.z.z.z]
INFO[0011] [hosts] Running cleaner container on host [z.z.z.z]
INFO[0012] [kube-cleaner] Successfully started [kube-cleaner] container on host [z.z.z.z]
INFO[0012] [hosts] Removing cleaner container on host [z.z.z.z]
INFO[0012] [hosts] Successfully cleaned up host [z.z.z.z]
INFO[0012] [etcd] Restoring [snapshot] snapshot on etcd host [x.x.x.x]
INFO[0013] [etcd] Successfully started [etcd-restore] container on host [x.x.x.x]
INFO[0014] [etcd] Restoring [snapshot] snapshot on etcd host [y.y.y.y]
INFO[0015] [etcd] Successfully started [etcd-restore] container on host [y.y.y.y]
INFO[0015] [etcd] Restoring [snapshot] snapshot on etcd host [z.z.z.z]
INFO[0016] [etcd] Successfully started [etcd-restore] container on host [z.z.z.z]
INFO[0017] [etcd] Building up etcd plane..
INFO[0018] [etcd] Successfully started [etcd] container on host [x.x.x.x]
INFO[0020] [etcd] Successfully started [rke-log-linker] container on host [x.x.x.x]
INFO[0021] [remove/rke-log-linker] Successfully removed container on host [x.x.x.x]
INFO[0022] [etcd] Successfully started [etcd] container on host [y.y.y.y]
INFO[0023] [etcd] Successfully started [rke-log-linker] container on host [y.y.y.y]
INFO[0025] [remove/rke-log-linker] Successfully removed container on host [y.y.y.y]
INFO[0025] [etcd] Successfully started [etcd] container on host [z.z.z.z]
INFO[0027] [etcd] Successfully started [rke-log-linker] container on host [z.z.z.z]
INFO[0027] [remove/rke-log-linker] Successfully removed container on host [z.z.z.z]
INFO[0027] [etcd] Successfully started etcd plane..
INFO[0027] Finished restoring on all etcd hosts
```

## Example

In this example we will assume that you started RKE on two nodes:

|  Name |    IP    |          Role          |
|:-----:|:--------:|:----------------------:|
| node1 | 10.0.0.1 | [controlplane, worker] |
| node2 | 10.0.0.2 | [etcd]                 |

### 1. Setting up rke cluster
A minimal cluster configuration file for running k8s on these nodes should look something like the following:

```
nodes:
  - address: 10.0.0.1
    hostname_override: node1
    user: ubuntu
    role: [controlplane,worker]
  - address: 10.0.0.2
    hostname_override: node2
    user: ubuntu
    role: [etcd]
```

After running `rke up` you should be able to have a two node cluster, the next step is to run few pods on node1:

```
kubectl --kubeconfig=kube_config_cluster.yml run nginx --image=nginx --replicas=3
```

### 2. Backup etcd cluster

Now lets take a snapshot using RKE:

```
rke etcd snapshot-save --name snapshot.db --config cluster.yml
```

![etcd snapshot](img/rke-etcd-backup.png)

### 3. Store snapshot externally

After taking the etcd backup on node2 we should be able to save this backup in a persistence place, one of the options to do that is to save the backup taken on a s3 bucket or tape backup, for example:

```
root@node2:~# s3cmd mb s3://rke-etcd-backup
root@node2:~# s3cmd /opt/rke/etcdbackup/snapshot.db s3://rke-etcd-backup/
```

### 4. Pull the backup on a new node

To simulate the failure lets powerdown node2 completely:

```
root@node2:~# poweroff
```

Now its time to pull the backup saved on s3 on a new node:

|  Name |    IP    |          Role          |
|:-----:|:--------:|:----------------------:|
| node1 | 10.0.0.1 | [controlplane, worker] |
| ~~node2~~ | ~~10.0.0.2~~ | ~~[etcd]~~                 |
| node3 | 10.0.0.3 | [etcd]                 |
|   |   |   |
```
root@node3:~# mkdir -p /opt/rke/etcdbackup
root@node3:~# s3cmd get s3://rke-etcd-backup/snapshot.db /opt/rke/etcdbackup/snapshot.db
```

### 5. Restore etcd on the new node

Now lets do a restore to restore and run etcd on the third node, in order to do that you have first to add the third node to the cluster configuration file:
```
nodes:
  - address: 10.0.0.1
    hostname_override: node1
    user: ubuntu
    role: [controlplane,worker]
#  - address: 10.0.0.2
#    hostname_override: node2
#    user: ubuntu
#    role: [etcd]
  - address: 10.0.0.3
    hostname_override: node3
    user: ubuntu
    role: [etcd]
```
and then run `rke etcd restore`:
```
rke etcd snapshot-restore --name snapshot.db --config cluster.yml
```

The previous command will restore the etcd data dir from the snapshot and run etcd container on this node, the final step is to restore the operations on the cluster by making the k8s api to point to the new etcd, to do that we run `rke up` again on the new cluster.yml file:
```
rke up --config cluster.yml
```
You can make sure that operations have been restored by checking the nginx deployment we created earlier:
```
> kubectl get pods                                                    
NAME                     READY     STATUS    RESTARTS   AGE
nginx-65899c769f-kcdpr   1/1       Running   0          17s
nginx-65899c769f-pc45c   1/1       Running   0          17s
nginx-65899c769f-qkhml   1/1       Running   0          17s
```

## License

Copyright (c) 2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
