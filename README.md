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

- [ ] Enable TLS for all subdomain
  ```
  An endpoint to download a wildcard certificate/key to use in your own server  
  ```
- [ ] Metrics

## Usage

```shell
go build
```