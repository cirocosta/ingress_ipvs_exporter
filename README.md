<h1 align="center">ingress_ipvs_exporter 📡  </h1>

<h5 align="center">Prometheus exporter for Docker Swarm Ingress IPVS metrics</h5>

<br/>

### Overview

`ingress_ipvs_exporter` is a [Prometheus](https://prometheus.io/) exporter focused on delivering statistics gathered via netlink regarding IPVS services that live inside the docker swarm's ingress network namespace.

```sh
Usage: ingress_ipvs_exporter 
	[--listen-address LISTEN-ADDRESS] 
	[--telemetry-path TELEMETRY-PATH] 
	[--namespace-path NAMESPACE-PATH]

Options:
  --listen-address LISTEN-ADDRESS
                         address to set the http server to listen to 
                         [default: :9100]

  --telemetry-path TELEMETRY-PATH
                         endpoint to receive scrape requests from prometheus 
                         [default: /metrics]

  --namespace-path NAMESPACE-PATH
                         absolute path to the network namespace where ipv is configured
                         [default: /var/run/docker/netns/ingress_sbox]

  --help, -h             display this help and exit
```

It exports the following metrics:


```
ipvs_bytes_in_total The total number of incoming bytes a virtual server
ipvs_bytes_out_total The total number of outgoing bytes from a virtual server
ipvs_connections_total The total number of connections made to a virtual server
ipvs_destination_active_connections_total The total number of connections established to a destination server
ipvs_destination_bytes_in_total The total number of incoming bytes to a real server
ipvs_destination_bytes_out_total The total number of outgoing bytes to a real server
ipvs_destination_connections_total The total number connections ever established to a destination
ipvs_destination_inactive_connections_total The total number of connections inactive but established to a destination server
ipvs_destination_total The total number of real servers that are destinations to the service
ipvs_services_total The total number of services registered in ipvs
```

Example:

```sh
# Create three services that publish ports in ingress
for i in $(seq 1 3); do 
	docker service create \
		--no-resolve-image \
		--detach=true \
		--name service_$i \
		--publish 80 \
		nginx:alpine
done

# Check the ports that these service have been bound to
docker service ls \
	--format '{{ json .Ports }}'
"*:30000->80/tcp"
"*:30001->80/tcp"
"*:30002->80/tcp"

# Make 10 requests to the first service
for i in $(seq 1 10); do
        curl \
                --silent \
                localhost:30000 \
                > /dev/null
done

# Start the exporter
sudo ingress_ipvs_exporter

# Check the metrics captured (stripped the other 
# services for better readability).
curl \
	--silent \
	localhost:9100/metrics | \
		ag ipvs

ipvs_bytes_in_total{fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 4510
ipvs_bytes_out_total{fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 11190
ipvs_connections_total{fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 10
ipvs_destination_active_connections_total{address="10.255.0.12",fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 0
ipvs_destination_bytes_in_total{address="10.255.0.12",fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 4510
ipvs_destination_bytes_out_total{address="10.255.0.12",fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 11190
ipvs_destination_connections_total{address="10.255.0.12",fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 10
ipvs_destination_inactive_connections_total{address="10.255.0.12",fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 10
ipvs_destination_total{fwmark="260",namespace="/var/run/docker/netns/ingress_sbox",port="30000"} 1
ipvs_services_total{namespace="/var/run/docker/netns/ingress_sbox"} 3
```


### Developing

Make sure you have the necessary permissions to run `modprobe`, `ip netns` and `ipvsadm`. 

Usually, that means that you need to execute `make test` as a superuser. 

Using `sudo`, make sure that `$PATH` is properly set - an easy way of doing so is modifying `/etc/sudoers` and adding the Go paths to the secure path.

