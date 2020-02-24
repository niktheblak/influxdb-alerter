import confuse
from datetime import datetime, timedelta
from dateutil import parser
from influxdb import InfluxDBClient
from pytz import timezone
import sys


def age(point):
    t = point['time']
    ts = parser.isoparse(t).replace(tzinfo=timezone('UTC'))
    now = datetime.utcnow().replace(tzinfo=timezone('UTC'))
    diff = now - ts
    return (ts, diff)


config = confuse.Configuration('influxdb-alerter', __name__)
max_age_cfg = config['max_age'].get(600)
max_age = timedelta(seconds=max_age_cfg)
client = InfluxDBClient(config['influxdb']['host'].get('localhost'),
                        config['influxdb']['port'].get(8086),
                        config['influxdb']['username'].get(str),
                        config['influxdb']['password'].get(str),
                        config['influxdb']['database'].get('ruuvitag'))
table = config['influxdb']['table'].get('ruuvitag_sensor')
result = client.query(f'SELECT * FROM "{table}" ORDER BY "time" DESC LIMIT 1')
if len(result) < 1:
    print("No results", file=sys.stderr)
    quit(1)
points = result.get_points(measurement=f"{table}")
for p in points:
    (ts, diff) = age(p)
    secs = diff.total_seconds()
    print(f"Last measurement was {secs} seconds ago at {ts}")
    if diff > max_age:
        print("No recent measurements", file=sys.stderr)
        quit(1)
points.close()
client.close()
