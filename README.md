# Aperture lnproxy

This project offers a dockerized solution for collecting fees on content delivery sites using the Lightning Network. It integrates Aperture, an HTTP 402 reverse proxy, and lnproxy, a wrapped invoice implementation, to enable seamless payment processing.

## Overview

### Sequence diagram

![image](https://github.com/motxx/aperture-lnproxy/assets/5776910/cf67a363-717c-4115-9dd7-175a7658e61b)

## Setup for lnproxy

* Add `./lnproxy/.env`

## Setup for aperture

* Add `./aperture/.env`

## Setup for contents

* Setup AWS S3 bucket (currently only public bucket supported)
* Add `./contents/.env`

### .lnd/ directory

Add `.lnd/` directory to project root:
```
.lnd
├── data
│   └── chain
│       └── bitcoin
│           └── mainnet
│               ├── admin.macaroon
│               └── invoice.macaroon
└── tls.cert
```

If you use Voltage Cloud, you can download `invoice.macaroon` by the following steps:
* Voltage Cloud > Manage Access > Macaroon Bakery
* Download `Type Invoice Default Invoice Macaroon`

### config/ directory

* Add `aperture.yaml` under the `config/` directory:
```
config
└── aperture.yaml
```

* Configure the following values in `aperture.yaml`:
  * `authenticator`
  * `servername`
    * e.g. `l402.example.com`
  * `services.hostregexp`
    * e.g. `l402.example.com`

### nginx/ directory

Add `ssl/` directory under the `nginx/` directory:
```
nginx
├── default.conf
└── ssl
    ├── fullchain.pem
    └── privkey.pem
```
`ssl/` directory includes TLS certifications for the hosting server domain (which is indicated as `l402.example.com` in `.example` files).

## Run docker compose

```
docker compose up -d
```

## Use CLI for remote content server

### Build CLI

```
cd contents
make build-cli
```

### Run CLI

```
appcli --rpcserver=l402.example.com:8080 \
  addcontent --id="avatar.png" --title="My Avatar" --author="moti" --filepath="under/the/s3/path/image.png" --recipient_lud16="moti@getalby.com" --price=30
```

```
appcli --rpcserver=l402.example.com:8080 \
  updatecontent --id="avatar.png" --title="My Avatar" --author="moti" --filepath="under/the/s3/path/image.png" --recipient_lud16="moti@getalby.com" --price=30
```

```
appcli --rpcserver=l402.example.com:8080 \
  removecontent --id="avatar.png"
```

```
appcli --rpcserver=l402.example.com:8080 \
  getcontent --id="avatar.png"
```
