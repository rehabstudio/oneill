oneill
======

oneill is a small tool that manages a set of docker containers running on a single host.


## How it works

- `oneill` loads a set of container definitions (configuration objects for
  each container that should be running) from a configured source (local
  directory/file, remote URL, etc.).
- container definitions are validated, any that don't pass validation are
  ignored.
- the latest image for each defined container is pulled from a remote registry
  (if possible), which can be either public or private.
- all running containers are checked to make sure that they:
  - match a valid definition (name, image, tag, env vars, etc)
  - are running the latest available image for the specified repository/tag
- containers that don't meet the criteria are removed and new containers are
  started so that a container is running for each valid container definition
- nginx configuration files are generated for each running container and a
  `SIGHUP` signal is sent to the server which causes it to perform a reread of
  all configuration files.


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
subdomain: example-subdomain     # required
image: example/some-container    # required
tag: v123                        # optional (default: "latest")
env:                             # optional (default: [])
  - "EXAMPLE=example"
  - "URL=http://www.example.com"
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
