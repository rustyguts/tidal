import subprocess
from pathlib import Path

import pytest

from tidal.tasks.concatenation import concatenate_chunks


class TestConcatenateChunks:
	def _make_chunks(self, temp_dir, ffmpeg_available, count=3):
		"""Helper to generate multiple video-only chunks for testing."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		Path(temp_dir).mkdir(parents=True, exist_ok=True)
		chunks = []
		for i in range(count):
			chunk_path = str(temp_dir / f"chunk_{i:04d}.mkv")
			result = subprocess.run(
				[
					"ffmpeg",
					"-y",
					"-f",
					"lavfi",
					"-i",
					f"testsrc2=size=320x240:rate=30:duration=0.5",
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

	def test_concatenate_chunks(self, temp_dir, ffmpeg_available):
		"""Test concatenating multiple video chunks."""
		chunks = self._make_chunks(temp_dir / "input", ffmpeg_available, count=3)
		output_dir = str(temp_dir / "output")

		result = concatenate_chunks.fn(
			chunk_paths=chunks,
			output_dir=output_dir,
			label="test",
		)

		assert Path(result).exists()
		assert "video_test" in result
		assert Path(result).stat().st_size > 0

	def test_concatenate_single_chunk(self, temp_dir, ffmpeg_available):
		"""Test concatenating a single chunk (degenerate case)."""
		chunks = self._make_chunks(temp_dir / "input", ffmpeg_available, count=1)
		output_dir = str(temp_dir / "output")

		result = concatenate_chunks.fn(
			chunk_paths=chunks,
			output_dir=output_dir,
			label="single",
		)

		assert Path(result).exists()

	def test_concatenate_empty_list_raises(self):
		"""Test that concatenating an empty list raises ValueError."""
		with pytest.raises(ValueError, match="No chunks"):
			concatenate_chunks.fn(
				chunk_paths=[],
				output_dir="/tmp/output",
				label="empty",
			)

	def test_concatenate_creates_output_dir(self, temp_dir, ffmpeg_available):
		"""Test that concatenation creates the output directory."""
		chunks = self._make_chunks(temp_dir / "input", ffmpeg_available, count=2)
		output_dir = str(temp_dir / "nested" / "output")

		result = concatenate_chunks.fn(
			chunk_paths=chunks,
			output_dir=output_dir,
			label="nested",
		)

		assert Path(result).exists()
		assert Path(output_dir).exists()

	def test_concatenate_cleans_up_concat_list(self, temp_dir, ffmpeg_available):
		"""Test that the temporary concat list file is cleaned up."""
		chunks = self._make_chunks(temp_dir / "input", ffmpeg_available, count=2)
		output_dir = str(temp_dir / "output")

		concatenate_chunks.fn(
			chunk_paths=chunks,
			output_dir=output_dir,
			label="cleanup",
		)

		# Check no concat_*.txt files remain in output_dir
		concat_files = list(Path(output_dir).glob("concat_*.txt"))
		assert len(concat_files) == 0
