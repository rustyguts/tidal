FROM python:3.13-alpine

RUN pip install -U uv

WORKDIR /app

ADD . /app

RUN uv venv .venv && uv sync --frozen