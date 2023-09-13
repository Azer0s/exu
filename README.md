# exu 🚞

A simple network emulation library for Go.

## Example

```go
sw1 := exu.NewEthernetSwitch("sw1", 10)

disconnectFn := func(p *exu.VPort) {
    sw1.DisconnectPort(p)
}

r1, _ := exu.NewRemoteVport(6554, net.ParseIP("10.0.0.1"), disconnectFn)
r2, _ := exu.NewRemoteVport(6555, net.ParseIP("10.0.0.2"), disconnectFn)

_ = sw1.ConnectToFirstAvailablePort(r1)
_ = sw1.ConnectToFirstAvailablePort(r2)

select {}
```

### On PC1

```bash
python3 -m http.server &
go run exu/client/main.go localhost 6554
```

### On PC2

```bash
go run exu/client/main.go localhost 6555
```

Now you can access `http://10.0.0.1:8000` from PC2.