# Content Discovery Service

### Install the app:

```
make install
```

### Diagrams:
*Basic App*
![](./images/app.jpg?raw=true "Basic App")

*Static Prices*
![](./images/aperture-static-prices.jpg?raw=true "Static Prices")

*Dynamic Prices*
![](./images/aperture-dynamic-prices.jpg?raw=true "Static Prices")


### CLI commands:

```
docker exec -it <container_name> \
  appcli addcontent --id="avatar.png" --title="My Avatar" --author="moti" --filepath="under/the/s3/path/image.png" --recipient_lud16="moti@getalby.com" --price=30
```

```
docker exec -it <container_name> \
  appcli updatecontent --id="avatar.png" --title="My Avatar" --author="moti" --filepath="under/the/s3/path/image.png" --recipient_lud16="moti@getalby.com" --price=30
```

```
docker exec -it <container_name> \
  appcli removecontent --id="avatar.png"
```

```
docker exec -it <container_name> \
  appcli getcontent --id="avatar.png"
```
