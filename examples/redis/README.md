oneill example configuration
============================

This example shows a single redis container with persistent storage. This
example uses the official redis image from the docker hub with
[RDB persistence](http://redis.io/topics/persistence). Port 6379 (the
standard redis port) is bound to the host interface and exposed.
