# exu 🚞

A simple network emulation library for Go.

## Example - Two PCs connected to a switch

```go
sw1 := exu.NewVSwitch("sw1", 10)

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

## Example - One PC connected to a router

```go
r1 := exu.NewVRouter("r1", 10)

r1 := exu.NewVRouter("r1", 10)
p1 := r1.GetFirstFreePort()
r1.SetPortIPNet(p1, net.IPNet{
    IP:   net.IPv4(10, 0, 0, 1),
    Mask: net.IPv4Mask(255, 255, 255, 0),
})

_, _ = exu.NewRemoteVport(6554, net.ParseIP("10.0.0.2"), func(port *exu.VPort) {
    _ = r1.ConnectPorts(p1, port)
}, func(p *exu.VPort) {
    log.Info("remote disconnected")
    r1.DisconnectPort(p1)
})

select {}
```

On PC1 you can run:

```bash
exu localhost 6554
```

And then ping the router:

```bash
ping -4 10.0.0.1
```