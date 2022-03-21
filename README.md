# vSCSI
iSCSI to vSock redirector 

# traffic flow

```mermaid
graph TD
    subgraph Guest
    A[iSCSI Client] --> B(iSCSI Driver)
    subgraph Kernel
    B --> C(TCP: 127.0.0.1:3260)
    subgraph Zero Copy
    C --> D(vSock Client: CID 2, Port 2222)
    end
    end
    end
    subgraph Host
    subgraph Zero Copy
    D --> E(vSock Server, CID 2, Port 2222)
    E --> F(TCP: 127.0.0.1:3260)
    end
    F --> G(iSCSI Server)
    end
```

# try it
## On host
1. Install & start iSCSI service with default port 3260
2. `go run main.go -s`

## In GuestOS
1. Install open-iscsi
2. `go run main.go`
3. `iscsiadm --mode discovery --type sendtargets --portal 127.0.0.1`  
