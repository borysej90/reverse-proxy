# Reverse Proxy

## How can someone build and run your code?
To build the reverse proxy, you need to run the following command from the repository root:
```shell
go build -o proxy ./cmd/proxy
```

Before you can run the binary, you have to set a YAML configuration file path to an environment variable like in the
example below:
```shell
export CONFIG_FILE=<path-to-yaml-config>
./proxy
```

You can also run the proxy directly without building it by running:
```shell
export CONFIG_FILE=<path-to-yaml-config>
go run ./cmd/proxy
```

### Tests
When making changes to the code, make sure all the tests pass. To run them, use the following command:
```shell
go test ./...
```
Additionally, you can provide a `-count N` flag to disable any test result caching and run tests N times. Sometimes
there are tests that can produce different results on each run but Go might cache the "successful" one and you will not
notice it. These are hard to debug test cases that are really dangerous.

### Configuration File
To run the proxy, you need to have a well-formed configuration file available. The schema is defined as a struct in
[config.go](internal/config/config.go). Currently, it supports data population from YAML files only. Here is an example
of a YAML file containing all the possible fields:
```yaml
server:
  listen: 80
  paths:
    - location: /
      target: http://internal.server.address
      connection_limit: 100
      drop_over_limit: true
```
This schema was inspired by Nginx's config documentation.

## What resources did you use to build your implementation?
I used two packages to simplify certain tasks:
- [YAML](https://github.com/go-yaml/yaml) package to unmarshal YAML configs
- I am used to [testify](https://github.com/stretchr/testify) package when writing assertions and mocks

Also, I referred to a built-in http package documentation and rate limiting patterns' Wiki page when developing the
reverse proxy.

## Explain any design decisions you made, including limitations of the system.
### Configuration
I have made the proxy configurable, including forwarding paths and targets, connection limits, and listening options. It
can be easily moved between environments, managed, and updated without the need of recompilation. You only need to
update the configuration file and restart the binary.

Speaking of configuration files, I have chosen YAML over, let's say, JSON because I find it more readable and
understandable. There are other options than config files, like environment variables, CLI parameters, or other
persistent storages, but the more parameters proxy accepts, the more verbose and unreadable it gets with, for instance,
CLI parameters or environment variables. Someone may say that these parameters can be scripted but config files still
benefit from structured, declarative format.

### Limiter
I have implemented the limiter as a `net/http` middleware - a `http.Handler` wrapper - that intercepts requests before
they are sent to the original handler and applies limiting policy. For the policy, I have implemented a Leaky Bucket
algorithm globally using a channel. It has a buffer size equal to the maximum number of concurrent requests to the origin
server.

Whenever a request arrives, proxy tries to send a `struct{}` to the channel, and if it is full, execution
blocks. After successful execution, proxy reads from the channel to decrement the number of elements in the channel and
allow other potentially blocked requests to continue execution. `struct{}` is an "empty struct" that literally has a
size of 0 bytes and every new instance points to the same value with no fields. This ensures that this channels requires
as little memory as possible.

### Limitations
The reverse proxy can only concatenate paths. For instance, if *location* is "/api/", *target* is
"ht<span>tp</span>://internal.server.address", and request URL is "/api/posts/123/", the proxy will make a request at
"ht<span>tp</span>://internal.server.address/api/posts/123/" address. However, one may want to strip the prefix that is
used to match against the target, i.e. "/api/", use a different prefix for target, or combination of both but it is not
possible in this iteration of the reverse proxy.

The proxy adds additional network and resource requirements, and possible latency with performance reduction. However,
this is not an implementation specific but rather general limitation that has to be considered whenever a reverse proxy
is used.

## How would you scale your implementation?
Even though this reverse proxy is not a load balancer, it cannot be horizontally scaled in a proper way. Horizontally
means running more copies of the proxy and putting a real load balancer in front of them. But this would not make a lot
of sense since each proxy would have a separate notion of number of concurrent requests to origin server.

However, usually reverse proxies are the load balancers and scaled vertically (increasing network bandwidth, updating hardware)
because you cannot just scale them horizontally and put another reverse proxy/load balancer in front of them since it
will inherently become a single server handling all the connections. Thus, I would stick to the vertical scaling
until it becomes impossible or unreasonable. Otherwise, there are only options of using load balancers on lower levels,
like transport layer, specialized hardware, or DNS round-robin.

## How would you make it more secure?
First, I would implement TLS support to enable HTTPS. This eliminates the possibility of the man-in-the-middle attacks.
Next, I would separate public and private networks such that only the reverse proxy is reachable from internet, while
all the servers are only accessible from the private subnet the proxy has access to. This reduces the possibility of
DDoS attacks, backdoor exploits, etc. on the origin servers.
