import subprocess
from pathlib import Path

import pytest

from tidal.flows.encode import encode_resolution
from tidal.models.transcode import CodecConfig, VideoResolution


@pytest.mark.integration
class TestEncodeResolutionFlow:
	def _make_chunks(self, temp_dir, count=2):
		"""Helper: generate video-only chunks for testing."""
		chunks = []
		temp_dir.mkdir(parents=True, exist_ok=True)
		for i in range(count):
			chunk_path = str(temp_dir / f"chunk_{i:04d}.mkv")
			result = subprocess.run(
				[
					"ffmpeg",
					"-y",
					"-f",
					"lavfi",
					"-i",
					"testsrc2=size=320x240:rate=30:duration=0.5",
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
				pytest.fail(f"Failed to create chunk: {result.stderr}")
			chunks.append(chunk_path)
		return chunks

	def test_encode_resolution_source(self, temp_dir, ffmpeg_available):
		"""Test encoding chunks at source resolution via full flow invocation."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		input_dir = temp_dir / "input"
		output_dir = str(temp_dir / "output")

		chunks = self._make_chunks(input_dir, count=2)
		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)

		# Call the flow directly (not .fn()) so Prefect sets up context
		result = encode_resolution(
			chunk_paths=chunks,
			resolution_label="240p",
			codec=codec,
			output_dir=output_dir,
		)

		assert Path(result).exists()
		assert Path(result).stat().st_size > 0

	def test_encode_resolution_with_scaling(self, temp_dir, ffmpeg_available):
		"""Test encoding chunks with downscaling via full flow invocation."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		input_dir = temp_dir / "input"
		output_dir = str(temp_dir / "output")

		chunks = self._make_chunks(input_dir, count=2)
		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)
		resolution = VideoResolution(width=160, height=120, label="120p")

		result = encode_resolution(
			chunk_paths=chunks,
			resolution_label="120p",
			codec=codec,
			output_dir=output_dir,
			resolution=resolution,
		)

		assert Path(result).exists()
		assert Path(result).stat().st_size > 0
