FROM python:3.9-rc AS builder

WORKDIR /app
RUN pip install --no-cache-dir poetry
ADD . .
RUN poetry build
RUN poetry export -f requirements.txt -o /app/dist/requirements.txt

FROM python:3.9-rc-slim-buster AS runner
COPY --from=builder /app/dist/* /dist/
RUN pip install --no-cache-dir -r /dist/requirements.txt
RUN pip install --no-index --find-links=file:///dist/ influxdb-alerter
VOLUME /root/.config/influxdb-alerter
CMD ["python", "-m", "influxdb_alerter"]
