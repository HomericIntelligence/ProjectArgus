FROM python:3.14.0-slim

# Single-stage: no pip installs or compiled extensions, so multi-stage adds no value.
RUN groupadd -r exporter && useradd -r -g exporter -u 1000 -m -s /sbin/nologin exporter

COPY --chown=exporter:exporter exporter.py /app/exporter.py

WORKDIR /app

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:9100/health')"

USER 1000

CMD ["python", "/app/exporter.py"]
