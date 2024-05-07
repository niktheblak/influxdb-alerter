FROM python:3.12-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
VOLUME /root/.config/influxdb-alerter

CMD ["python", "-m", "influxdb_alerter"]
