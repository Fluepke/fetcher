TODO: Write README

You need an IPv6 prefix, a router that allows static route configuration and you need to patch `main.go`

1. Enable IPv6 non local bind: `sudo sysctl -w net.ipv6.ip_nonlocal_bind=1`
2. Configure a static route to that host. On you router run e.g.: `sudo ip route add 2a0f:5382:1312:8000::/49 via 2a0f:5381::103`
3. Put packets destined to the attack network into local packet processing. On the attack machine run e.g. `sudo ip route add to local 2a0f:5382:1312:8000::/49 dev enp1s0`