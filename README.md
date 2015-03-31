# gerrycode [![Build Status](https://travis-ci.org/paked/go-repo-review.svg)](https://travis-ci.org/paked/go-repo-review)
A project designed to help developers find the best open source projects to contribute to, or depend on.

#Generate your keys!
In order for JWTs to work, you need a public and a private key. Here are some simple steps to do this

1. ```openssl genrsa -out app.rsa 1024```
2. ```openssl rsa -in app.rsa -pubout > app.rsa.pub```

You can either pass in paths to these keys using the ```-public``` and ```-private``` flags, or place them in the keys directory.

##This project is **not** finished

