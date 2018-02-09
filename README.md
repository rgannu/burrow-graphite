# burrow-graphite

A simple graphite exporter for gathering Kafka consumer group info
from [burrow](https://github.com/linkedin/Burrow).

The source code base is from prometheus [burrow-exporter](https://github.com/jirwin/burrow_exporter) and it has been adapted for graphite support.

## Run with Docker

### required environment variables
#### BURROW_ADDR
A burrow address is required. Default: http://localhost:8000
#### GRAPHITE_HOST
A graphite host to connect. Default: 192.168.99.100
#### GRAPHITE_PORT
Graphite port. Default: 2003
#### INTERVAL
A scrape interval is required. Default: 30

### Example
```sh
# with env variables
docker run --add-host=dh:<burrow-ip> \
  -e BURROW_ADDR="http://dh:8000" \
  -e GRAPHITE_HOST="<graphite-server-host>" \ 
  -e GRAPHITE_PORT="2003" \
  -e INTERVAL="30" \
  <image-id>

# with custom command
docker run --add-host=dh:<burrow-ip> \ 
  <image-id> ./burrow-graphite \
  --burrow-addr http://dh:8000 \
  --graphite-host <graphite-server-host> \
  --graphite-port 2003 \
  --interval 30
```
