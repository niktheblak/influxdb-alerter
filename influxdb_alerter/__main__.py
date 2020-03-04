import logging
import smtplib
from datetime import datetime, timedelta
from pathlib import Path
from email.message import EmailMessage

import confuse
import pytz
from dateutil import parser
from influxdb import InfluxDBClient


def point_ts(point):
    t = point['time']
    return parser.isoparse(t).replace(tzinfo=pytz.utc)


def age(ts):
    now = datetime.utcnow().replace(tzinfo=pytz.utc)
    return now - ts


def send_email(cfg, last_measurement_ts):
    email_cfg = cfg['email']
    with smtplib.SMTP(email_cfg['host'].get(str), email_cfg['port'].get(465)) as s:
        msg = EmailMessage()
        msg.set_charset('UTF-8')
        msg.set_content(
            f"There has been no new measurements in InfluxDB since {last_measurement_ts.ctime()}.")
        msg['Subject'] = "No new InfluxDB measurements"
        msg['From'] = email_cfg['from'].get(str)
        msg['To'] = email_cfg['to'].get(str)
        s.starttls()
        s.login(email_cfg['username'].get(str),
                email_cfg['password'].get(str))
        s.send_message(msg)


logging.basicConfig(level=logging.INFO)
config = confuse.Configuration('influxdb-alerter', __name__)
max_age = timedelta(seconds=config['max_age'].get(600))
email_enabled = config['email']['enabled'].get(True)
email_sent_file = config['email']['sent_file'].as_filename()
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
ts = point_ts(p)
diff = age(ts)
local_ts = ts.astimezone()
logging.info(
    "Last measurement was %s seconds ago at %s",
    diff.total_seconds(),
    local_ts.ctime())
path = Path(email_sent_file)
if diff > max_age:
    logging.warning("No new measurements since %s", local_ts.ctime())
    if email_enabled and not path.exists():
        send_email(config, local_ts)
        path.touch()
else:
    path.unlink(missing_ok=True)
