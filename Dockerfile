FROM python:3.13.5-slim

RUN apt update && apt install -y wget xz-utils curl unzip && rm -rf /var/lib/apt/lists/*

RUN wget -q https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-n7.1-latest-linux64-gpl-7.1.tar.xz && \
	tar -xvf ffmpeg-n7.1-latest-linux64-gpl-7.1.tar.xz && \
	mv ffmpeg-n7.1-latest-linux64-gpl-7.1/bin/ffmpeg /usr/local/bin/ffmpeg && \
	mv ffmpeg-n7.1-latest-linux64-gpl-7.1/bin/ffprobe /usr/local/bin/ffprobe && \
	rm -rf ffmpeg-n7.1-latest-linux64-gpl-7.1 && \
	rm ffmpeg-n7.1-latest-linux64-gpl-7.1.tar.xz

RUN pip install -U uv

WORKDIR /app

ADD . /app

RUN uv venv .venv && uv sync --frozen