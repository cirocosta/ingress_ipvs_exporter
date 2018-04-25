<h1 align="center">ipvs_exporter 📡  </h1>

<h5 align="center">Prometheus exporter for IPVS metrics</h5>

<br/>

### Developing

Make sure you have the necessary permissions to run `modprobe`. 

Usually, that means that you need to execute `make test` as a superuser. Using `sudo`, make sure that `$PATH` is properly set - an easy way of doing so is modifying `/etc/sudoers` and adding the Go paths to the secure path.

