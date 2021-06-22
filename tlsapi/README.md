# tlsapi

Support local development with proper HTTPS configuration without dealing with manually trusting self-signed CA certificates.

## Important/Disclaimer

⚠️ This tool doesn't installs a root CA in your system but will require server to serve using a pre-made certificate. Use it only if you know what you are doing.


## Features

- [x] Support TLS for all subdomains (requires your server to serve with our certificate) 
  ```  
  https://1-2-3-4.devns.net
  https://2400-cb00-2049-1--a29f-1804.devns.net      
  https://localhost.devns.net
  ```

- [ ] Auto-renew certificate (let's encrypt only provides 3-months certificate)

- [ ] Protection against malicious certificate revocation 

- [ ] 3rd party integration guide for production use
  
- [ ] Metrics

## Usage

### CLI

```shell
Usage of ./tlsapi:
  -domain string
        [MUST CHANGE] Base domain for DNS resolution (default "example.com")
  -h    Print this help
  -port int
        Port for DNS server to listen to (default 8888)
  -tls
        Turn on TLS mode
  -tls-dryrun
        Set to use STAGING ACME Directory
  -tls-email string
        [MUST CHANGE] Email for letsencrypt registration (default "john.doe@example.com")
```

### Get certificate

TODO