# exu 🚞

A simple vSwitch/VPN in Go.

## Testing

```bash
python3 -m http.server
```

```bash
exu server
```

```bash
exu client
```

Now you can access `http://localhost:8080` via the VPN IP address (something in the 10.0.0.1/24 range).