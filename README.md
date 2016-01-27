# mqtt-listener
connects to mqtt, listens for events and sends them to time series db
# TODO
* enable humidity sending to influx
* fix inability to stay connected to mqtt 
** somehow stops receiving updates
* enable multiple sensors configuration 
** currently only one at a time is supported

# Future Enhancements
* binary mqtt message format instead of json