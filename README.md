# Reverse Proxy

## How can someone build and run your code?
To build the reverse proxy, you must run the following command from the repository root:
```shell
go build -o proxy ./cmd/proxy
```

Before running the binary, you must set a YAML configuration file path to an environment variable, like in the example
below:
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
Additionally, you can provide a `-count N` flag to turn off any test result caching and run tests N times. Sometimes
some tests can produce different results on each run, but Go might cache the "successful" one, and you will not notice
it. These are hard-to-debug test cases that are really dangerous.

### Configuration File
To run the proxy, you must have a well-formed configuration file available. The schema is defined as a struct in
config.go. Currently, it supports data population from YAML files only. Here is an example of a YAML file containing all
the possible fields:
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
I used two packages to simplify specific tasks:
- [YAML](https://github.com/go-yaml/yaml) package to unmarshal YAML configs
- I am used to [testify](https://github.com/stretchr/testify) package when writing assertions and mocks

Also, I referred to a built-in http package documentation and rate limiting patterns' Wiki page when developing the
reverse proxy.

## Explain any design decisions you made, including limitations of the system.
### Configuration
I have made the proxy configurable, including forwarding paths and targets, connection limits, and listening options.
You can easily move it between environments, manage, and update it without recompilation. You only need to update the
configuration file and restart the binary.

Speaking of configuration files, I have chosen YAML over, let's say, JSON because I find it more readable and
understandable. There are other options than config files, like environment variables, CLI parameters, or other
persistent storage. However, the more parameters the proxy accepts, the more lengthy and unreadable it gets with, for
instance, CLI parameters or environment variables. Some say these parameters can be scripted, but config files still
benefit from a structured, declarative format.

### Limiter
I have implemented the limiter as a `net/http` middleware - an `http.Handler` wrapper - that intercepts requests before
they are sent to the original handler and applies a limiting policy. For the policy, I have implemented a Leaky Bucket
algorithm on a per-server basis using channels. It has a buffer size equal to the maximum number of concurrent requests
to the origin server.

Whenever a request arrives, the proxy tries to send a `struct{}` to the channel, and if it is full, the execution
blocks. After successful execution, the proxy reads from the channel to decrease the number of elements in it and allow
other potentially blocked requests to continue execution. `struct{}` is an "empty struct" that has a size of 0 bytes,
and every new instance points to the same value with no fields. It ensures that these channels require as little memory
as possible.

### Limitations
The reverse proxy can only concatenate paths. For instance, if *location* is "/api/", *target* is
"ht<span>tp</span>://internal.server.address", and the request URL is "/api/posts/123/", the proxy will request
"ht<span>tp</span>://internal.server.address/api/posts/123/" address. However, one may want to strip the prefix used to
match against the target, i.e., "/api/", use a different prefix for the target, or a combination of both, but it is not
possible in this iteration of the reverse proxy.

The proxy adds additional network and resource requirements and possible latency with performance reduction. However,
this is not an implementation-specific but rather a general limitation that has to be considered whenever a reverse
proxy is used.

## How would you scale your implementation?
Even though this reverse proxy is not a load balancer, it cannot be scaled horizontally in a proper way. Horizontally
means running more proxy copies and putting an actual load balancer before them. However, this would make little sense
since each proxy would have a separate notion of the current number of concurrent requests to the origin server.

However, usually, reverse proxies are the load balancers. They are scaled vertically (increasing network bandwidth,
updating hardware) because you cannot just scale them horizontally and put another reverse proxy/load balancer in front
of them since it will inherently become a single server handling all the connections. Thus, I would stick to the
vertical scaling until it becomes impossible or unreasonable. Otherwise, there are only options for using load balancers
on lower levels, like transport layer, specialized hardware, or DNS round-robin.

## How would you make it more secure?
First, I would implement TLS support to enable HTTPS. It eliminates the possibility of the man-in-the-middle attacks.
Next, I would separate public and private networks such that only the reverse proxy is reachable from the internet. At
the same time, all the servers are only accessible from the private subnet the proxy has access to. It reduces the
possibility of DDoS attacks, backdoor exploits, etc., on the origin servers.
