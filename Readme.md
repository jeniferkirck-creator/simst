# SIMST 
The client is designed for testing web applications and services under extreme network load conditions. If the configuration is correct, it will attempt to maximize server load with transport-layer and application-layer requests. To increase the load, you can run multiple instances on a single host or distribute clients across multiple hosts.<br><br>
The client makes highly efficient use of the host's CPU and RAM. Periodically check the device's status.
> Attention!
> 
> Externally, this testing will look like a DDoS-attack on the server. Do not run the client against any service without explicit authorization from its owner. Unauthorized testing is unethical and illegal in most countries! We strongly condemn such practices!

---

### Requirments

To build the source code, you will need [golang](https://go.dev/dl/) 1.25 or later and any code editor of your choice. The client can be built for any operating system and architecture, although Linux is recommended. At the transport layer, all requests are sent via a raw socket with specific settings, so this functionality will not work on other operating systems. On Linux-like systems such as BusyBox, we also cannot guarantee that this part of the functionality will work. 

---

### Usage

To run the client,specify the path to the global configuration file with the -c flag. If omitted, the client will look for config.json in the same directory as the client executable. See examples/global/ for an example file.

Configuration parameters: 

* target_file_patch - path to the file containing the list of test requests (optional)
* test_duration - test duration in minutes
* request_timeout - delay between request cycles in microseconds
* stepping_payload - ramp-up load settings (not yet implemented, optional)
* server - server access settings (optional)

  * target_link - path for retrieving the test request list
  * report_link -  path for sending the stress test report
  * register_link -  path for client registration
* result_file_patch -  path to the file for writing test results
* public_ip - public IP (optional)

The client supports two modes: local and server. To enable server mode, provide the server access settings and leave the request list file path empty or omitted. See examples/configuration/ for an example JSON request configuration file. 

---

###  Control Server

The server can be written in any language you like. For proper operation, it must handle the requests described below.

1. **Registration**:<br>Request : HTTP/1 GET http://host.com/register?ip=<br>Response: status code
2. **Configuration**:<br>Request : HTTP/1 GET http://host.com/configuration<br>Response: JSON, fully matching the structure shown in the example
3. **Results**:<br>Request : HTTP/1 POST http://host.com/report Request body:<br>

```json 
    {
      "timestamp": 1775788726, 
      "requests_count": 0,
      "payload_length": 0,
      "codes": [200,200,201,500]
   }
```

Response: status code

---

We hope this client will help you improve your service and prepare it for high loads!

