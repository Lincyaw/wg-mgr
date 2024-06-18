
# Debian

Prepare:

```bash
sudo apt install wireguard
```

Usage:

1. Update server config

```bash
# case 1
./vpn-tool setup > /etc/wireguard/wg0.conf
wg syncconf wg0 <(wg-quick strip wg0) 

# case 2
./vpn-tool setup > /etc/wireguard/wg0.conf
wg-quick down wg0 # optional
wg-quick up wg0 
```

2. Add user

```bash
./vpn-tool adduser --id yourname
```

2.1 If want to advertise a subnet router, 

```bash
./vpn-tool adduser --id router \
    --postup "iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -s 100.10.10.0/24 -o eth0 -j MASQUERADE" \
    --postdown "iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -s 100.10.10.0/24 -o eth0 -j MASQUERADE" \
    --advertise-routes "10.10.10.0/24"
```

2.2 To add a user that accepts the routes,

```bash
./vpn-tool adduser --id client --accept-routes
```

3. Delete user

```bash
./vpn-tool deluser --id yourname
```

4. Delete user

```bash
./vpn-tool getuser --id yourname
```
