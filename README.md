# dnsserver

Support local development with proper HTTPS configuration without dealing with self-signed certificates 

## Features 

- [x] Resolve IP domain into IP
    ```HOSTS
    1-2-3-4.ipv4.iptls.com 1.2.3.4
    2400-cb00-2049-1--a29f-1804.ipv6.iptls.com 2400:cb00:2049:1::a29f:1804 
    localhost.ipv4.iptls.com 127.0.0.1
    localhost.ipv6.iptls.com ::1 
    ```

- [x] Support TLS certificate generation
  ```
  Generate a wildcard certificate to use in your own server  
  ```
- [ ] Metrics

## Usage

```shell
Usage of ./dnsserver:
  -domain string
        [MUST CHANGE] Base domain for DNS resolution (default "example.com")
  -h    Print this help
  -nameserver string
        [MUST CHANGE] Primary NS for SOA must end with period(.) (default "ns.example.com.")
  -port int
        Port for DNS server to listen to (default 53)
  -soa-email string
        Email for SOA must end with period(.) (default "john\\n.doe.example.com.")
  -tls
        Turn on TLS mode
  -tls-dryrun
        Set to use STAGING ACME Directory
  -tls-email string
        [MUST CHANGE] Email for letsencrypt registration (default "john.doe@example.com")
```