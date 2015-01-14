oneill
======

oneill is a small tool that manages a set of docker containers running on a single host.


## How it works

- `oneill` reads configuration from a flat folder structure, 1 directory per
  configured site.
    - `siteconfig.yaml`: main configuration file for container/site
    - `.env`: encrypted environment variables that will be decrypted and
      passed to the container at runtime #TODO
- A yaml configuration file is parsed and details of all containers that
  should be running are loaded
- container definitions are validated, any that don't pass validation are
  ignored in any future steps
    - checks for duplicate subdomains
    - checks subdomain is a valid string (a-z, A-Z, 0-9, -_)
    - check that the specified container/tag exists and can be pulled locally
    - check that the container exposes a single port
- stop and remove any containers that don't match one of our valid definitions
- start containers for all valid definitions (if they're not already running)
- remove all existing nginx configurations
- generate a new nginx config for every running container
- issue an `nginx reload` which performs a hot reload of all configurations
  (and won't result in any downtime)


## Configuration structure

Configuration is managed by a collection of yaml files in a structured
directory. Users can optionally add an encrypted `.env` file that will be
decrypted and passed to the container at runtime.


```
.
+-- demo-golang
|   +-- siteconfig.yaml
|   +-- .env
+-- my-nodejs-site
|   +-- siteconfig.yaml
+-- python-flask-example
|   +-- siteconfig.yaml
```

Eack `siteconfig.yaml` file *must* contain the following two keys:

```yaml
subdomain: example-subdomain
container: example/some-container
```


## Installation

Download the release binary and make it executable.

```bash
$ wget http://storage.googleapis.com/rehab-labs-oneill-releases/stable/oneill
$ chmod +x oneill
```


## Usage

By default `oneill` will look for its configuration file at
`/etc/oneill/config.yaml`, but you can specify an alternate location using the
`-config=` command line flag if you wish.

You should take care to only run one oneill process at a time (for now,
this'll be safer later). The easiest way to accomplish this is with the
`run-one` tool on ubuntu.

```bash
$ apt-get install run-one
$ ./oneill -config=myconfig.yaml
