FROM python:3.13-slim AS builder

ARG POETRY_INSTALL_FLAGS=""

ENV POETRY_VERSION=2.1.1
RUN pip install -U "poetry==$POETRY_VERSION"

WORKDIR /app

COPY poetry.lock pyproject.toml /app/

RUN poetry config virtualenvs.in-project true && \
  poetry install --no-interaction --no-root --no-ansi $POETRY_INSTALL_FLAGS

ENV PATH="/app/.venv/bin:$PATH"
COPY . /app/

FROM python:3.13-slim

WORKDIR /app
COPY --from=builder /app/.venv /app/.venv
ENV PATH="/app/.venv/bin:$PATH"
COPY . /app/