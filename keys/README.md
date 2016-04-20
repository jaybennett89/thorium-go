# How to Generate RSA Keys

While in this folder, execute the following commands to generate new RSA keys on Ubuntu 14.04

```
openssl genrsa -out app.rsa 2048
openssl rsa -in app.rsa -pubout > app.rsa.pub
```

For more details, see the following link: https://gist.github.com/cryptix/45c33ecf0ae54828e63b
