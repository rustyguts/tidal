import json
import os
import shutil
import subprocess
import tempfile
from pathlib import Path
from unittest.mock import MagicMock

import pytest


@pytest.fixture(scope="session")
def ffmpeg_available() -> bool:
	"""Check if ffmpeg is available on the system."""
	return shutil.which("ffmpeg") is not None


@pytest.fixture(scope="session")
def ffprobe_available() -> bool:
	"""Check if ffprobe is available on the system."""
	return shutil.which("ffprobe") is not None


@pytest.fixture
def temp_dir(tmp_path):
	"""Provide a temporary directory for test outputs."""
	return tmp_path


@pytest.fixture(scope="session")
def sample_video(tmp_path_factory, ffmpeg_available) -> str:
	"""Generate a short (2-second) test video with audio using FFmpeg's test sources.

	Returns the path to the generated video file. This fixture is session-scoped
	to avoid regenerating the video for every test.
	"""
	if not ffmpeg_available:
		pytest.skip("FFmpeg not available")

	video_dir = tmp_path_factory.mktemp("sample_video")
	video_path = str(video_dir / "test_video.mp4")

	result = subprocess.run(
		[
			"ffmpeg",
			"-y",
			"-f",
			"lavfi",
			"-i",
			"testsrc2=size=320x240:rate=30:duration=2",
			"-f",
			"lavfi",
			"-i",
			"sine=frequency=440:duration=2",
			"-c:v",
			"libx264",
			"-preset",
			"ultrafast",
			"-crf",
			"28",
			"-pix_fmt",
			"yuv420p",
			"-c:a",
			"aac",
			"-b:a",
			"64k",
			video_path,
		],
		capture_output=True,
		text=True,
		timeout=30,
	)

	if result.returncode != 0:
		pytest.fail(f"Failed to generate sample video: {result.stderr}")

	return video_path


@pytest.fixture(scope="session")
def sample_video_no_audio(tmp_path_factory, ffmpeg_available) -> str:
	"""Generate a short test video without audio."""
	if not ffmpeg_available:
		pytest.skip("FFmpeg not available")

	video_dir = tmp_path_factory.mktemp("sample_video_no_audio")
	video_path = str(video_dir / "test_video_no_audio.mp4")

	result = subprocess.run(
		[
			"ffmpeg",
			"-y",
			"-f",
			"lavfi",
			"-i",
			"testsrc2=size=320x240:rate=30:duration=2",
			"-c:v",
			"libx264",
			"-preset",
			"ultrafast",
			"-crf",
			"28",
			"-pix_fmt",
			"yuv420p",
			"-an",
			video_path,
		],
		capture_output=True,
		text=True,
		timeout=30,
	)

	if result.returncode != 0:
		pytest.fail(f"Failed to generate sample video: {result.stderr}")

	return video_path


@pytest.fixture(scope="session")
def sample_audio(tmp_path_factory, ffmpeg_available) -> str:
	"""Generate a short (2-second) test audio file."""
	if not ffmpeg_available:
		pytest.skip("FFmpeg not available")

	audio_dir = tmp_path_factory.mktemp("sample_audio")
	audio_path = str(audio_dir / "test_audio.m4a")

	result = subprocess.run(
		[
			"ffmpeg",
			"-y",
			"-f",
			"lavfi",
			"-i",
			"sine=frequency=440:duration=2",
			"-c:a",
			"aac",
			"-b:a",
			"64k",
			audio_path,
		],
		capture_output=True,
		text=True,
		timeout=30,
	)

	if result.returncode != 0:
		pytest.fail(f"Failed to generate sample audio: {result.stderr}")

	return audio_path


@pytest.fixture(scope="session")
def sample_chunk(tmp_path_factory, ffmpeg_available) -> str:
	"""Generate a short video-only chunk (no audio) for encoding tests."""
	if not ffmpeg_available:
		pytest.skip("FFmpeg not available")

	chunk_dir = tmp_path_factory.mktemp("sample_chunk")
	chunk_path = str(chunk_dir / "test_chunk.mkv")

	result = subprocess.run(
		[
			"ffmpeg",
			"-y",
			"-f",
			"lavfi",
			"-i",
			"testsrc2=size=320x240:rate=30:duration=1",
			"-c:v",
			"libx264",
			"-preset",
			"ultrafast",
			"-crf",
			"28",
			"-pix_fmt",
			"yuv420p",
			"-an",
			chunk_path,
		],
		capture_output=True,
		text=True,
		timeout=30,
	)

	if result.returncode != 0:
		pytest.fail(f"Failed to generate sample chunk: {result.stderr}")

	return chunk_path


@pytest.fixture
def mock_ffmpeg_processor(mocker):
	"""Mock the FFmpegProcessor for unit tests that don't need real FFmpeg."""
	mock_processor_class = mocker.patch("tidal.utilities.ffmpeg.FFmpegProcessor")
	mock_instance = MagicMock()
	mock_processor_class.return_value = mock_instance
	mock_instance.execute_sync.return_value = {
		"returncode": 0,
		"stdout": "",
		"stderr": "",
		"command": ["ffmpeg", "-y"],
	}
	return mock_instance


@pytest.fixture
def sample_probe_json() -> dict:
	"""Return sample ffprobe JSON output for mocking."""
	return {
		"streams": [
			{
				"index": 0,
				"codec_type": "video",
				"codec_name": "h264",
				"width": 1920,
				"height": 1080,
				"r_frame_rate": "30/1",
				"bit_rate": "5000000",
			},
			{
				"index": 1,
				"codec_type": "audio",
				"codec_name": "aac",
				"sample_rate": "48000",
				"channels": 2,
			},
		],
		"format": {
			"duration": "120.5",
			"bit_rate": "5500000",
			"format_name": "mov,mp4,m4a,3gp,3g2,mj2",
		},
	}
