# keyvalue

This project is a very simple key-value store implemented in Go. It allows you to store and retrieve arbitrary key-value pairs in a persistent log file. It supports basic operations like Set, Get, Delete, and Compact (for cleaning up deleted/outdated entries). The store can be configured to optionally cache data in memory for faster reads or only in a persistant log file, and it offers basic safeguards like key and value size limits.

## Disclaimer

I created this project with a very specific use case in mind. I make a lot of simple web applications that don't need a fully-featured database and don't get a lot of traffic. My goal was to create a very simple library that can be used to store small amounts of arbitrary data, without the need for any external dependencies. If this is something that interests you, this project may be a good fit. It's much more likely, however, that you'd be better off using a proper database or key-value store.
