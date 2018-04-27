<h1 align="center">ipvs_exporter 📡  </h1>

<h5 align="center">Prometheus exporter for IPVS metrics</h5>

<br/>

### Overview

`ipvs_exporter` is a Prometheus exporter focused on delivering statistics gathered via netlink regarding IPVS services that live inside network namespaces.

```
Usage: ipvs_exporter [--listen-address LISTEN-ADDRESS] [--telemetry-path TELEMETRY-PATH] --namespace-path NAMESPACE-PATH

Options:
  --listen-address LISTEN-ADDRESS
                         address to set the http server to listen to [default: :9100]
  --telemetry-path TELEMETRY-PATH
                         endpoint to receive scrape requests from prometheus [default: /metrics]
  --namespace-path NAMESPACE-PATH
                         absolute path to the network namespace where ipv is configured
  --help, -h             display this help and exit
```

### Developing

Make sure you have the necessary permissions to run `modprobe`, `ip netns` and `ipvsadm`. 

Usually, that means that you need to execute `make test` as a superuser. 

Using `sudo`, make sure that `$PATH` is properly set - an easy way of doing so is modifying `/etc/sudoers` and adding the Go paths to the secure path.

