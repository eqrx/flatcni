# flatcni

## What is this

The usual Kubernets experience gives you an an isolated cluster network with CoreDNS as an internal DNS 
server and kube-proxy for load balancer tooling. While that is great for general setups, the original author 
of this likes to have his globally routable IPv6 addresses that are resolvable by public DNS everywhere, so he
ripped out CoreDNS and kube-proxy and developed a custom CNI plugin and operators.

## Components

All of the components are bundled inside a single binary.

### flatcni

A CNI plugin that assigns pods a globally routable IPv6 address.

### flatcni-dns

Operator that watches EndpointSlices inside a cluster and updates corresponding DNS records via RFC2136.
(Yes, external-dns exists but author found it to bloated).

### flatcni-firewall

Operator that watches EndpointSlices inside a cluster and allows inbound traffic to services.

## License

This project is licensed under AGPL-3.0 License (sse LICENSE file).
