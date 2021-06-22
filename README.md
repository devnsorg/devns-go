# devns

Support local development

- proper HTTPS configuration without dealing with manually trusting self-signed CA certificates.
- Tunneling using Wireguard

## Important/Disclaimer

⚠️ This tool doesn't installs a root CA in your system but will require server to serve using a pre-made certificate. Use it only if you know what you are doing.


## Features

- [x] Resolve IP domain into IP using `dnsserver` module
  ```
  # HOSTS
  1-2-3-4.devns.net 1.2.3.4
  2400-cb00-2049-1--a29f-1804.devns.net 2400:cb00:2049:1::a29f:1804 
  localhost.devns.net 127.0.0.1
  localhost.devns.net ::1 
  ```

- [x] Support TLS for all subdomains using `tlsapi` module (requires your server to serve with our certificate) 
  ```  
  https://1-2-3-4.devns.net
  https://2400-cb00-2049-1--a29f-1804.devns.net      
  https://localhost.devns.net
  ```
  
- [ ] Support tunneling using Wireguard
 
- [ ] Metrics

## Usage

TODO