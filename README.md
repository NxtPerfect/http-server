# HTTP server
simple http 1.0 server implementation in go using "net" library

# Performance
Each request requested either GET for 308B html file (index.html) or POST for 33.59KiB html file (form.html)

Each connection was used for one request, and immediately closed after getting a response

Each response was tested for valid output

1 000 000 requests took 485.4s on 8c 16t cpu

2060 requestes per second

485.4μs per request


# Features
GET/POST requests are accepted
returns html files
custom http paths
