
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

3. Delete user

```bash
./vpn-tool deluser --id yourname
```

4. Delete user

```bash
./vpn-tool getuser --id yourname
```
