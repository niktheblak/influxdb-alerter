FROM python:3.8-buster

RUN pip install --no-cache-dir poetry
ADD pyproject.toml ./
ADD poetry.lock ./
RUN poetry install --no-dev --no-root
ADD . .
VOLUME /root/.config/influxdb-alerter
ENTRYPOINT ["poetry", "run", "python", "alerter.py"]
