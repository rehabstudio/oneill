oneill
======

oneill is a small tool that manages a set of docker containers running on a
single host, exposing them to the internet using Nginx as a reverse proxy.


## Why another container orchestration tool?

oneill is a much more narrowly focussed, more opinionated tool than most
others. It's designed to fulfil a very specific requirement, running a set of
12-factor style application containers on a single host based on a simple
yaml/json configuration format.

oneill expects to take full control of a host server on which both docker and
nginx are already running and installed. oneill loads its configuration from
one of a number of configurable sources (single file, directory of multiple
files, remote HTTP API etc.), pulls the required images from public or private
registries, starts/stops the required containers with configured environment
variables and reconfigures nginx to point the right subdomains at the right
applications.

## What it doesn't do

oneill does not provide any support for persistence, applications/containers
must use a remote backing store such as Amazon S3 or Google Cloud Storage.
Filesystems within containers *are* writable, but should be considered
ephemeral as they may be thrown away if oneill decides to restart/upgrade the
container for any reason.

oneill is not a suitable platform for containers which need durable
persistence (at present, this may change in future) like database servers or
anything which couldn't cope with losing all of its state at any moment.
Depending on your workload oneill may be a suitable platform for cache servers
like memcached or redis.

oneill doesn't concern itself with clustering or running multiple hosts, but
it's built in way that makes running in those situations as simple as possible
(just point ansible/puppet/chef at an extra host, no extra setup).


## What is it useful for?

oneill may have a narrow focus and opinionated feature set, but it's quite
useful in a number of scenarios and for a number of reasons. The following
list is just an example of the situations in which oneill is useful:

- As a one-off deploy step to pull the latest image for a container, start it
  up and perform a zero downtime upgrade of the application.
- Run on a schedule (maybe via cron) to continuously check a registry for
  updates and perform a zero-downtime update when one is found.
- Run on multiple servers with the same configuration (possibly behind a load
  balancer) to easily scale out a set of containers across multiple hosts.
- A continuous integration server pushing successful builds to a private
  docker registry could easily be combined with oneill to form part of a
  continuous delivery workflow.
- oneill's support for reading container definitions from a remote HTTP API
  could be used to form the building blocks of a PaaS-like system. oneill
  can easily be made to work with something like `etcd` (or your own custom
  API) as a control channel.
- Providing a consistent deployment and operations platform, regardless of the
  type of application being built or stack being used (provided it conforms to
  the 12-factor style, of course).


## How it works

oneill follows a simple, linear flow for the most part, aiming to be easy to
understand without any magic. The best way to understand what oneill is doing
is to dive into the code (starting at `main.go` it should be pretty
straightforward I hope), but a rough overview is provided below:

```
    Load configuration from disk
    Load container definitions (from disk/remote api/etc)
    Validate container definitions
    For each valid container definition:
        Pull latest docker image (if available)
        Validate docker image
        Check if a container is already running tha matches the definition
        Start a new container if necessary
    Write new nginx configuration and htpasswd files to disk
    Remove redundant nginx configuration and htpasswd files
    Reload nginx configuration
    Sleep for a few seconds to allow old connections to finish gracefully
    Stop and remove old/redundant docker containers
```


## Installation

Only Linux builds are available at present, but mac/windows builds should be
possible, I just don't have those platforms to test on.

Download the latest binary and make it executable (preferably somewhere on
your `PATH` like `/usr/local/bin` but it doesn't really matter). Although the
url below has the word "stable" in it, releases should be considered alpha
quality at best until we say otherwise (Stable releases will be available once
we have something to release).

```bash
$ wget http://storage.googleapis.com/rehab-labs-oneill-releases/stable/oneill
$ chmod +x oneill
```


## Configuration

oneill configuration is managed by a single yaml/json file which is read from
the filesystem at runtime. The default location for oneill's configuration
file is `/etc/oneill/config.yaml`, but this can be overridden on the command
line with the `-config=` flag (example in the "Usage" section below).

See `example.config.yaml` for an explanation of the available settings and all
possible values.


## Container Definitions

Container definitions define exactly what containers should be running on the
system and in what state. oneill ensures that the containers running on the
system match what is defined by the container definitions.

oneill provides several options for loading container definitions. How the
definitions are loaded is controlled via a URI specified in the configuration
file. The following types of URI are supported:

- `file:///path/to/some/directory`: The given directory is scanned for all
  files with the extensions `.yaml` and `.json`. Each file should contain a
  single container definition at the top level (i.e. one definition per file).
- `file:///path/to/some/file.yaml`: The given file is read from disk and
  parsed as JSON/YAML. The file should contain a list (array) of container
  definitions (multiple containers per file).
- `https://www.somedomain.com/api/that/returns/json/or/yaml/`: The remote URL
  is fetched and parsed as JSON/YAML. The response should contain a list
  (array) of container definitions.

Definitions can also be read from STDIN. When data is passed to STDIN, any URI
specified in the configuration file will be ignored. For example:

```bash
$ cat containers.yaml | oneill
```

A single container definition should contain the following data:

```yaml
# subdomain that this container will be served on (also used to form part of
# the running container name). This value is required.
subdomain: example-subdomain

# the docker image that will be used to run this container. This value is
# required.
image: example/some-container

# the specific tag that should be used when running this container. This value
# is optional (default: "latest").
tag: v123

# disable nginx globally (disable all interaction with nginx). This value is
# optional (default: false).
nginx_disabled: false

# When an image exposes more than one port, the following option must be set
# to a valid port number that's exposed by the image. If this is not set
# correctly oneill will not be able to expose this container via nginx. This
# value is optional (default: 0).
nginx_exposed_port: 0

# add custom environment variables that will be passed into the container when
# started. This value is optional (default: []).
env:
  - "EXAMPLE=example"
  - "URL=http://www.example.com"

# adding htpasswd entries will cause oneill to lock this container/subdomain
# down behind HTTP basic auth. Note: unless you're running on HTTPS it
# probably isn't a good idea to use this feature.. This value is optional
# (default: []).
htpasswd:
  - bob:$apr1$SBA9z0lK$B7c8xGmNJ427sINH2BGEr.
  - jon:$apr1$SBA9z0lK$B7c8xGmNJ427sINH2BGEr.
```


## Usage

oneill has a single command line option, most settings are only available via
the config file, so running is simple:

```bash
# run oneill with the default config file (/etc/oneill/config.yaml)
$ oneill

# run oneill with a custom config file
$ oneill -config=/home/me/my_oneill_config.yaml
```


## Building from source

oneill uses `go-bindata` to embed some assets into the built binary, you'll
need to have this installed before you're able to build the application.

oneill uses `godep` to manage its dependencies. Provided you have `godep`
installed and the repository is cloned into your `$GOPATH`, building oneill
locally should be as simple as:

```bash
$ go-bindata -o nginxclient/bindata.go -pkg=nginxclient -prefix=nginxclient/ nginxclient/templates/
$ godep go build
```


## Building with docker

If you've got docker installed and are running Linux (it might work on OSX, i
just haven't tried), you can use the included script to build (and test)
oneill:

```bash
$ ./build.sh docker
```
