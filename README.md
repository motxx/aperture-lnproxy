# Aperture lnproxy

This project offers a dockerized solution for collecting fees on content delivery sites using the Lightning Network. It integrates Aperture, an HTTP 402 reverse proxy, and lnproxy, a wrapped invoice implementation, to enable seamless payment processing.

## Overview

### Sequence diagram

![image](https://github.com/motxx/aperture-lnproxy/assets/5776910/cf67a363-717c-4115-9dd7-175a7658e61b)

### Implementation summary

* Fork lnproxy and add aperture
* Implement a new challenger for aperture that supports lnproxy
* Request lnurlp to get creator invoice
* Implement docker compose files

## Setup for lnproxy

## Setup for aperture

### .lnd/ directory

* `./.lnd/tls.cert`
* `./.lnd/data/chain/bitcoin/mainnet/invoice.macaroon`

If you use Voltage Cloud, you can download `invoice.macaroon` by the following steps:
* Voltage Cloud > Manage Access > Macaroon Bakery
* Download `Type Invoice Default Invoice Macaroon`

### config/ directory

* `./config/aperture.yaml`

### nginx/ directory

Add TLS certifications for the hosting server domain (which is indicated as `l402.example.com` in `.example` files):
* `./nginx/ssl/fullchain.pem`
* `./nginx/ssl/privkey.pem`

## Run docker compose

```
docker compose up -d
```
