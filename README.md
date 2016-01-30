# mqtt-listener
connects to mqtt, listens for events and sends them to time series db

# TODO
* enable multiple sensors configuration (B1820 and DHT22 simultainously)
** currently only one at a time is supported
** on different mqtt paths
* make influks.go a service ran at startup
* add SSL and client authentication
    or hide it behind VPN


# Future Enhancements
* binary message format instead of json
* switch from mqtt to gRPC / thrift?