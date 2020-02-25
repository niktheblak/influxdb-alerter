import logging
import smtplib
from datetime import datetime, timedelta
from pathlib import Path

import confuse
import pytz
from dateutil import parser
from influxdb import InfluxDBClient


def age(point):
    t = point['time']
    ts = parser.isoparse(t).replace(tzinfo=pytz.utc)
    now = datetime.utcnow().replace(tzinfo=pytz.utc)
    diff = now - ts
    return ts, diff


def send_email(cfg, last_measurement_ts):
    with smtplib.SMTP_SSL(cfg['smtp']['host'].get(str), cfg['smtp']['port'].get(465)) as s:
        s.ehlo()
        s.login(cfg['smtp']['username'].get(str),
                cfg['smtp']['password'].get(str))
        from_email = cfg['smtp.from'].get(str)
        to_email = cfg['smtp.to'].get(str)
        msg = f"There has been no new measurements in InfluxDB since {last_measurement_ts.ctime()}."
        s.sendmail(from_email, to_email, msg)


logging.basicConfig(level=logging.INFO)
config = confuse.Configuration('influxdb-alerter', __name__)
max_age = timedelta(seconds=config['max_age'].get(600))
email_sent_file = config['email_sent_file'].as_filename()
client = InfluxDBClient(config['influxdb']['host'].get('localhost'),
                        config['influxdb']['port'].get(8086),
                        config['influxdb']['username'].get(str),
                        config['influxdb']['password'].get(str),
                        config['influxdb']['database'].get('ruuvitag'))
table = config['influxdb']['table'].get('ruuvitag_sensor')
result = client.query(f'SELECT * FROM "{table}" ORDER BY "time" DESC LIMIT 1')
if len(result) < 1:
    logging.error("No results")
    quit(1)
points = result.get_points(measurement=f"{table}")
p = next(points)
points.close()
client.close()
(ts, diff) = age(p)
local_ts = ts.astimezone()
secs = diff.total_seconds()
logging.info(
    "Last measurement was %s seconds ago at %s",
    secs,
    local_ts.ctime())
path = Path(email_sent_file)
if diff > max_age:
    logging.warning("No new measurements since %s", local_ts.ctime())
    if not path.exists():
        send_email(config, local_ts)
        path.touch()
else:
    path.unlink(missing_ok=True)
