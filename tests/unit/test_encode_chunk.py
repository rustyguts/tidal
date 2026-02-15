from pathlib import Path

import pytest

from tidal.models.transcode import CodecConfig, VideoResolution
from tidal.tasks.encode_chunk import encode_chunk


class TestEncodeChunk:
	def test_encode_chunk_source_resolution(self, sample_chunk, temp_dir, ffmpeg_available):
		"""Test encoding a chunk at source resolution."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)

		result = encode_chunk.fn(
			chunk_path=sample_chunk,
			output_dir=str(temp_dir),
			codec=codec,
			chunk_index=0,
			resolution_label="source",
		)

		assert Path(result).exists()
		assert "encoded_source_0000" in result
		assert Path(result).stat().st_size > 0

	def test_encode_chunk_with_scaling(self, sample_chunk, temp_dir, ffmpeg_available):
		"""Test encoding a chunk with resolution scaling."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)
		resolution = VideoResolution(width=160, height=120, label="120p")

		result = encode_chunk.fn(
			chunk_path=sample_chunk,
			output_dir=str(temp_dir),
			codec=codec,
			chunk_index=0,
			resolution=resolution,
			resolution_label="120p",
		)

		assert Path(result).exists()
		assert "encoded_120p_0000" in result

	def test_encode_chunk_creates_output_dir(self, sample_chunk, temp_dir, ffmpeg_available):
		"""Test that encoding creates the output directory if it doesn't exist."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)
		nested_dir = temp_dir / "nested" / "output"

		result = encode_chunk.fn(
			chunk_path=sample_chunk,
			output_dir=str(nested_dir),
			codec=codec,
			chunk_index=5,
			resolution_label="source",
		)

		assert Path(result).exists()
		assert "0005" in result

	def test_encode_chunk_different_codecs(self, sample_chunk, temp_dir, ffmpeg_available):
		"""Test encoding with different codec configurations."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		configs = [
			CodecConfig(video_codec="libx264", preset="ultrafast", crf=23),
			CodecConfig(video_codec="libx264", preset="ultrafast", crf=35),
		]

		for i, codec in enumerate(configs):
			result = encode_chunk.fn(
				chunk_path=sample_chunk,
				output_dir=str(temp_dir / f"codec_{i}"),
				codec=codec,
				chunk_index=0,
				resolution_label=f"test_{i}",
			)
			assert Path(result).exists()

	def test_encode_chunk_numbering(self, sample_chunk, temp_dir, ffmpeg_available):
		"""Test that chunk numbering in output filename is correct."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(video_codec="libx264", preset="ultrafast", crf=28)

		for idx in [0, 42, 999]:
			result = encode_chunk.fn(
				chunk_path=sample_chunk,
				output_dir=str(temp_dir / f"chunk_{idx}"),
				codec=codec,
				chunk_index=idx,
				resolution_label="source",
			)
			assert f"{idx:04d}" in Path(result).name
