# dnsserver

Support local development with proper HTTPS configuration without dealing with manually trusting self-signed CA certificates.

## Important/Disclaimer

⚠️ This tool doesn't installs a root CA in your system but will require server to serve using a pre-made certificate. Use it only if you know what you are doing.


## Features

- [x] Resolve IP domain into IP
  ```
  # HOSTS
  1-2-3-4.iptls.com 1.2.3.4
  2400-cb00-2049-1--a29f-1804.iptls.com 2400:cb00:2049:1::a29f:1804 
  localhost.iptls.com 127.0.0.1
  localhost.iptls.com ::1 
  ```

- [x] Support TLS for all subdomains (requires your server to serve with our certificate) 
  ```  
  https://1-2-3-4.iptls.com
  https://2400-cb00-2049-1--a29f-1804.iptls.com      
  https://localhost.iptls.com
  ```

- [ ] Auto-renew certificate (let's encrypt only provides 3-months certificate)

- [ ] Protection against malicious certificate revocation 

- [ ] 3rd party integration guide for production use
  
- [ ] Metrics

## Usage

### CLI

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

### Get certificate

TODO