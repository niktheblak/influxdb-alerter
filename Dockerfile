FROM python:3.9-slim-buster

WORKDIR /app
ADD . .
RUN pip install --no-cache-dir -r requirements.txt
VOLUME /root/.config/influxdb-alerter
CMD ["python", "-m", "influxdb_alerter"]
