# Aperture lnproxy

コンテンツへの支払のL402 paywallであるapertureと、Wrapped invoiceのlnproxyを組み合わせたdocker composeのコンテナです。コンテンツの料金をクリエイターに支払う一方で、金額の一部をコンテンツ配信サービスを運用する事業者が、手数料として徴収できるようにします。

## Disclaimer

L402とWrapped Invoiceを組み合わせる本リポジトリの方法は、L402の権利となるpreimageを確率的に不正入手する手段が潜在しています。現実的にどの程度不正が可能かは検証できておりません。詳しくは下記をご覧ください。
https://coinkeninfo.com/wrapped-invoice/

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

If you use Voltage Cloud, you can download `admin|invoice.macaroon` by the following steps:
* Voltage Cloud > Manage Access > Macaroon Bakery
* Download `Type Admin|Invoice Default Invoice Macaroon`

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
make daemon
```

## Use CLI for remote content server

### Build CLI

```
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
