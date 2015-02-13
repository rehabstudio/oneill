oneill example configuration
============================

This example shows a set of 4 worker processes running in containers, 2 that
require persistence, 2 that don't. The worker containers are loaded from a
private registry, with custom credentials.

Note: The containers in this example aren't real, but you can pretend they're
some sort of celery/sidekiq workers.
