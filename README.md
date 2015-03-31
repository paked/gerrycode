# gerrycode [![Build Status](https://travis-ci.org/paked/gerrycode.svg?branch=master)](https://travis-ci.org/paked/gerrycode)
Gerrycode is a website helping you get (and give) feedback on your code and projects.

#Generate your keys!
In order for JWTs to work, you need a public and a private key. Here are some simple steps to do this

1. ```openssl genrsa -out app.rsa 1024```
2. ```openssl rsa -in app.rsa -pubout > app.rsa.pub```

You can either pass in paths to these keys using the ```-public``` and ```-private``` flags, or place them in the keys directory.

## Status
**Frontend:** Useable, not done. (60%)

**Backend/API:** Almost finished. (90%)
