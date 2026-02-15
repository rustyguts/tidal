FROM python:3.14-slim

RUN apt update && apt install -y wget xz-utils curl unzip && rm -rf /var/lib/apt/lists/*

# Install FFmpeg 7.1 with VMAF support (GPL build includes libvmaf)
# Detect architecture: Docker sets TARGETARCH to "amd64" or "arm64"
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
		FFMPEG_ARCH="linuxarm64"; \
	else \
		FFMPEG_ARCH="linux64"; \
	fi && \
	FFMPEG_TAR="ffmpeg-n7.1-latest-${FFMPEG_ARCH}-gpl-7.1.tar.xz" && \
	wget -q "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/${FFMPEG_TAR}" && \
	tar -xf "${FFMPEG_TAR}" && \
	FFMPEG_DIR=$(basename "${FFMPEG_TAR}" .tar.xz) && \
	mv "${FFMPEG_DIR}/bin/ffmpeg" /usr/local/bin/ffmpeg && \
	mv "${FFMPEG_DIR}/bin/ffprobe" /usr/local/bin/ffprobe && \
	rm -rf "${FFMPEG_DIR}" "${FFMPEG_TAR}"

RUN pip install -U uv

WORKDIR /app

# Copy dependency and build files first for better layer caching
COPY pyproject.toml uv.lock README.md ./

# Install dependencies only (skip installing the project itself since source isn't copied yet)
# --extra dev includes watchfiles (needed for hot-reload in docker-compose)
RUN uv venv .venv && uv sync --frozen --no-install-project --extra dev

# Copy application code
COPY tidal/ ./tidal/
COPY deploy.py ./

# Now install the project itself with source available
RUN uv sync --frozen --extra dev
