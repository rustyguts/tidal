import pytest

from tidal.models.transcode import ProbeResult
from tidal.tasks.segmentation import segment_video


class TestSegmentVideo:
	def test_segment_real_video(self, sample_video, temp_dir, ffmpeg_available):
		"""Test segmenting a real video produces chunks and extracts audio."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		result = segment_video.fn(
			source_path=sample_video,
			work_dir=str(temp_dir),
			segment_duration=1,
		)

		assert result.segment_count > 0
		assert len(result.chunk_paths) == result.segment_count
		assert result.audio_path is not None
		assert result.work_dir == str(temp_dir)

		# Verify chunk files exist
		from pathlib import Path

		for chunk_path in result.chunk_paths:
			assert Path(chunk_path).exists()

		# Verify audio file exists
		assert Path(result.audio_path).exists()

	def test_segment_video_no_audio(self, sample_video_no_audio, temp_dir, ffmpeg_available):
		"""Test segmenting a video without audio."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		probe = ProbeResult(
			duration=2.0,
			width=320,
			height=240,
			video_codec="h264",
			frame_rate=30.0,
			has_audio=False,
		)

		result = segment_video.fn(
			source_path=sample_video_no_audio,
			work_dir=str(temp_dir),
			segment_duration=1,
			probe=probe,
		)

		assert result.segment_count > 0
		assert result.audio_path is None

	def test_segment_nonexistent_file(self, temp_dir):
		"""Test that segmenting a nonexistent file raises FileNotFoundError."""
		with pytest.raises(FileNotFoundError, match="does not exist"):
			segment_video.fn(
				source_path="/nonexistent/video.mp4",
				work_dir=str(temp_dir),
			)

	def test_segment_creates_work_dir(self, sample_video, temp_dir, ffmpeg_available):
		"""Test that segmentation creates the work directory structure."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		work_dir = temp_dir / "nested" / "work"

		result = segment_video.fn(
			source_path=sample_video,
			work_dir=str(work_dir),
			segment_duration=1,
		)

		from pathlib import Path

		chunks_dir = Path(work_dir) / "chunks"
		assert chunks_dir.exists()
		assert result.segment_count > 0

	def test_segment_short_duration(self, sample_video, temp_dir, ffmpeg_available):
		"""Test segmenting with a duration longer than the video produces one chunk."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		result = segment_video.fn(
			source_path=sample_video,
			work_dir=str(temp_dir),
			segment_duration=30,  # Longer than our 2s test video
		)

		assert result.segment_count == 1
