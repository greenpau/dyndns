# dyndns

<a href="https://github.com/greenpau/dyndns/actions/" target="_blank"><img src="https://github.com/greenpau/dyndns/workflows/build/badge.svg?branch=master"></a>

Dynamic DNS Registrator for Route 53

## Getting Started

First, create AWS CloudFormation stack named `DynDnsUpdateServiceUser` with
`assets/cloudformation/dyndns_service_user.yaml`.

Next, create `~/dyndns_config.json` configuration file:

```json
{
  "provider": {
    "type": "route53",
    "zone_id": "Z627GH1M87Y192",
    "credentials": "~/.aws/credentials",
    "profile_name": "dyndns"
  },
  "record": {
    "name": "app.contoso.com",
    "type": "A",
    "ttl": 60
  },
  "sync_interval": 60
}
```

Finally, start the `dyndns` service:

```bash
bin/dyndns --config ~/dyndns_config.json --log-level debug
```

## Deployment

First, install `dyndns`:

```
sudo yum -y localinstall dyndns-1.0.1-1.el7.x86_64.rpm
```

Then, amend the following files:

* `/etc/dyndns/config.json`
* `/var/lib/dyndns/.aws/credentials`

After that, enable and start the service:

```bash
sudo systemctl enable dyndns
sudo systemctl start dyndns
sudo systemctl status dyndns
sudo journalctl -u dyndns -r --no-pager | head -100
```
