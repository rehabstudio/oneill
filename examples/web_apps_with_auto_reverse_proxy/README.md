oneill example configuration
============================

This example shows a simple set of container definitions, 4 web applications
and an nginx container acting as a reverse proxy. This example assumes that
`A` DNS records for `*.mydomain.com` and `*.myotherdomain.com` have been
created and pointed at the server that oneill is running on.


## Web Applications

This example contains 4 simple web applications, a collection of "hello world"
demos in various languages/frameworks: Python, Go, NodeJS and Ruby. The URL
each application is served on is controlled by an environment variable
configured in each container definition.


## Reverse Proxy

This example includes an automated reverse proxy, built using nginx and
docker-gen. [nginx-proxy](https://github.com/jwilder/nginx-proxy/) sets up a
container running nginx and docker-gen. docker-gen generates reverse proxy
configs for nginx and reloads nginx when containers are started and stopped.

You can find more details on how to configure nginx-proxy in its
[README](https://github.com/jwilder/nginx-proxy/).
