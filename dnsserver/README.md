# dnsserver

Support local development with proper DNS resolution for IP address

## Features

- [x] Resolve IP domain into IP
  ```
  # HOSTS
  1-2-3-4.devns.net 1.2.3.4
  2400-cb00-2049-1--a29f-1804.devns.net 2400:cb00:2049:1::a29f:1804 
  localhost.devns.net 127.0.0.1
  localhost.devns.net ::1 
  ```
 
- [ ] Metrics

## Usage

### CLI

```shell
Usage of ./dnsserver:
  -api-endpoint value
        Specify multiple value for calling multiple API to get result
  -domain string
        [MUST CHANGE] Base domain for DNS resolution (default "example.com")
  -h    Print this help
  -nameserver string
        [MUST CHANGE] Primary NS for SOA must end with period(.) (default "ns.example.com.")
  -port int
        Port for DNS server to listen to (default 53)
  -soa-email string
        Email for SOA must end with period(.) (default "john\\n.doe.example.com.")
```
