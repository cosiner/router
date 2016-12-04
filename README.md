# Router

Router is a url router library for [Go](https://golang.org).

# Documentation
Documentation can be found at [Godoc](https://godoc.org/github.com/cosiner/router).

# Syntax
* *Static*: `/user`
* *Param*: `/user/:id` will catch user id, `/user/:/following` catch nothing but do only matching.
* *CatchAny*: `/folder/*path` will catch any characters after `/folder/`, there should be nothing after it.
* *Param* and *CatchAny* can has a optional regexp flag, just like `/user/:id:[\da-f]+` and `/folder/*path:.*\.go`.
* The matching order is: *Static*, *Param*(with regexp flag), *Param*, *CatchAny*(with regexp flag), *CatchAny*.

* **NOTE**: `/` match only `/`, `/*` match anything.
      
# LICENSE
MIT.
