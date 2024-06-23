# Telsa BLE Proxy

A stripped down version of a Tesla BLE proxy made for use with evcc.

## How to use

1. [Create a key](https://github.com/teslamotors/vehicle-command/blob/main/cmd/tesla-control/README.md) as described in the Tesla API documentation.

2. Create a file called "tesla" in the directory where you want to run tesla-proxy from:

```bash
export TESLA_VIN=<YOUR-VIN>
export TESLA_KEYFILE=/<path-to-key>/private.pem
```

3. Install the service: (Edit the service file first to match your paths)

```bash
sudo cp tesla-proxy.service /etc/systemd/system/
sudo systemctl enable tesla-proxy
sudo systemctl start tesla-proxy
```