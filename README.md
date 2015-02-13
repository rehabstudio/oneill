oneill
======

oneill is a small tool that manages a set of docker containers running on a
single host. It uses a simple YAML/JSON configuration format to define the
containers that should running.


## Why another container orchestration tool?

oneill is a smaller, more opinionated tool than most others. It's designed to
perform a single simple role, running a configured set of docker containers on
a single host. oneill is designed to be the glue layer between a loosely
coupled set of components which run in docker containers.

It's not accurate to call oneill a PaaS, but it's designed to provide a great
foundation for building your own bespoke PaaS-like infrastructure.

oneill has only a single requirement, that docker is installed and running,
additionally, it expects to have full control of *all* containers running on
the system on which it is installed. Be careful, if you run oneill it will
remove any containers not defined in its own configuration.

oneill loads its container definitions from one of a number of configurable
sources (single file, directory of multiple files, remote HTTP API etc.), the
definitions are then validated for correctness. oneill will pull any required
images from the appropriate public or private registries each time it runs,
ensuring the latest image is always available. Containers will then be
stopped/started/upgraded as required to ensure that the running set matches
the validated container definitions loaded earlier.


## Networking

By default, all ports exposed by a container will be mapped to the host
interface on a random port (the equivalent of `docker run -P`). You are highly
encouraged to use the default settings when exposing ports as docker will take
care of conflicts for you automatically.

If you absolutely need to expose a particular container on a specific port
(maybe you want to run nginx on ports 80 and 443) then oneill allows you map
specific ports when defining a container. See the example container definition
below (or browse the examples directory) for a fuller explanation.


## Persistence

By default, containers do not have any guarantee of persistence. Whilst
volumes are writable during the lifetime of a container, unless persistence
support is explicitly enabled in the container definition all volumes will be
removed when a container is restarted, removed or upgraded.

oneill maintains its own folder hierachy in a configurable location for
mounting volumes into containers. The only reason oneill does not just let
docker manage the location of the data is that this method makes it easier to
persist data across container restarts/changes.

It's important to note that oneill does not allow mounting of arbitrary
directories or files into a container, it simply persists the volumes already
defined in the image being used.


## Interacting with docker from within a container

If configured to do so, oneill can bind-mount the unix socket used for the
docker API in a container so it can be accessed by applications running within
it. It is not reccommended to enable this feature for most containers but it
can be very useful for building certain types of services. Some simple
examples of where this feature can be used:

- Watch for changes to running containers and configure a reverse proxy to
  route HTTP traffic to properly configured applications.
- Expose system stats by reading from the Docker stats API and presenting via
  a simple web dashboard.
- Connect to stdout/stderr of other running containers and persist log data
  off-server.
- Announce container information via a service-discovery system like etcd.


## What is it useful for?

oneill may have a narrow focus and opinionated feature set, but it's quite
useful in a number of scenarios and for a number of reasons. The following
list is just an example of some of the situations in which oneill is useful:

- As a one-off deploy step to pull the latest image for a container, and run
  it with a specific configuration.
- Run on a schedule (maybe via cron) to continuously check a registry for
  updates and perform appropriate updates when any are found.
- Run on multiple servers with the same configuration (possibly behind a load
  balancer) to easily scale out a set of 12-factor style containers across
  multiple hosts.
- A continuous integration server pushing successful builds to a private
  docker registry could easily be combined with oneill to form part of a
  continuous delivery workflow.
- oneill's support for reading container definitions from a remote HTTP API
  could be used to form the building blocks of a PaaS-like system.


## How it works

oneill follows a simple, linear flow for the most part, aiming to be easy to
understand without any magic. The best way to understand what oneill is doing
is to dive into the code (starting at `main.go` it should be pretty
straightforward I hope), but a rough overview is provided below:

```
    Load configuration from disk
    Load container definitions (from disk/remote api/etc)
    Validate container definitions
    Stop and remove old/redundant docker containers
    For each valid container definition:
        Pull latest docker image (if available)
        Validate docker image
        Check if a container is already running that matches the definition
        Stop the old container if necessary
        Start a new container if necessary
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
- `stdin://`: oneill will load container definitions passed via STDIN, e.g.
  `cat containers.yaml | oneill`


## Example container definition

A number of example configurations and definition setups are included in the
examples directory, you can browse those to get an idea of how oneill can be
used.

A single container definition should contain the following data:

```yaml
# container_name controls the user-specified part of the name oneill will give
# to the container at startup time. This setting is required.
container_name: example-name

# repo_tag controls the container that will be pulled and run for this
# container definition. This is in the same format as you would pass to
# `docker run`, e.g. `locahost:5000/myimage:latest`, `nginx`, `ubuntu:14.04`,
# `my.private.repo/myotherimage`. This setting is required.
repo_tag: example/some-container
```

In addition to the required settings above, the follwing optional settings can
also be added to a container definition.

```yaml
# add custom environment variables that will be passed into the container when
# started. This value is optional (default: []).
env:
  - "EXAMPLE=example"
  - "URL=http://www.example.com"

# should persistence be enabled for this container? default off as we don't
# want to encourage people to use local persistence (whilst acknowledging that
# it is necessary in some situations).
persistence_enabled: false

# should the docker control socket be bind-mounted into this container? this
# is useful for service containers that need to be able to see or control what
# other containers are doing (automated logging, reverse proxy, etc. need this
# functionality).
docker_control_enabled: false

# service containers allow an explicit port mapping as some services need to
# be exposed on specific ports to be useful e.g. nginx on 80/443 for serving
# http. Regular containers do not need this functionality. Keys are host port
# numbers and values are the internal port numbers that should be exposed.
port_mapping:
  80: 80
  443: 443
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

oneill uses `godep` to manage its dependencies. Provided you have `godep`
installed and the repository is cloned into your `$GOPATH`, building oneill
locally should be as simple as:

```bash
$ godep go build
```


## Building with docker

If you've got docker installed and are running Linux (it might work on OSX, i
just haven't tried), you can use the included script to build (and test)
oneill:

```bash
$ ./build.sh docker
```
