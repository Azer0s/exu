# exu 🚞

A simple network emulation library for Go.

## Example

```go
sw1 := exu.NewEthernetSwitch("sw1", 10)

disconnectFn := func(p *exu.VPort) {
    sw1.DisconnectPort(p)
}

connectFn := func(port *exu.VPort) {
    _ = sw1.ConnectToFirstAvailablePort(port)
}

go exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), connectFn, disconnectFn)
go exu.NewRemoteVport(6555, net.ParseIP("10.0.0.2"), connectFn, disconnectFn)

select {}
```

### On PC1

```bash
python3 -m http.server &
go run exu/client/main.go localhost 6554
```

### On PC2

```bash
export pc1_ip="address of PC1"
go run exu/client/main.go $pc1_ip 6555
```

Now you can access `http://10.0.0.1:8000` from PC2.