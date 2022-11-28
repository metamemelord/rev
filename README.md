# rev

Rev  is tunneling tool which directs TCP traffic from one machine to other. Rev can redirect traffic from public endpoints to local machine.

### Usage
#### Build the binary
```sh
go build -o rev main.go
```

#### Run as server
```sh
rev server -p <port>
```

#### Run as client and connect to a running server
```sh
rev connect -d <destination-post> -u <user-name> -n <service-name> -s <server-host> -p <server-port>
```

#### Example
Start **Server** running on host example.com on port 3000
```
rev server -p 3000
```

**External service running on https://localhost:5001**

Connecting to rev server to redirect traffic from server to external service
```sh
rev connect -u myusername -n myservicename -s example.com -p 3000 --destination-protocol https -d 5001
```

All the traffic coming to https://example.com:3000/messages/myusername/myservicename will be redirected to service running at https://localhost:5001.
