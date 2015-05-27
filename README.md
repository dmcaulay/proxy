# proxy

The proxy uses consistent hashing to map statsd metric names to statsd instances.

## Project Status

This code has not been tested in a production environemnt. I wrote this while I was learning golang as an experiment. I wanted to have a simple networking app that I could compare with a similar Node.js project. You can find the Node.js version at https://github.com/etsy/statsd/blob/master/proxy.js.

## Install

```
$ go get github.com/dmcaulay/proxy
```

## Test

```
$ godep go test
```

## Configuration

The configuration files are stored within the `config` directory and can be specified via the environment parameter.

```
$ proxy -e=production
```

The file format is json and you can find the production config at `config/production.json`

```js
{
  "Nodes": [
    {"Host": "127.0.0.1", "Port": 8127},
    {"Host": "127.0.0.1", "Port": 8129},
    {"Host": "127.0.0.1", "Port": 8131}
  ],
  "UdpVersion": "udp4",
  "Host":  "0.0.0.0",
  "Port": 8125
}
```

The `Nodes` attribute specifies the statsd instances and the `UdpVersion`, `Host` and `Port` attributes specify the proxy configuration.

## Run

```
$ godep go install
$ proxy
```
