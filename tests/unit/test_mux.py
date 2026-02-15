from pathlib import Path

import pytest

from tidal.tasks.mux import mux_audio_video


class TestMuxAudioVideo:
	def test_mux_audio_video(self, sample_video_no_audio, sample_audio, temp_dir, ffmpeg_available):
		"""Test muxing audio and video into a final output."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		result = mux_audio_video.fn(
			video_path=sample_video_no_audio,
			audio_path=sample_audio,
			output_dir=str(temp_dir),
			label="test",
			container="mp4",
		)

		assert Path(result).exists()
		assert "final_test.mp4" in result
		assert Path(result).stat().st_size > 0

	def test_mux_mkv_container(self, sample_video_no_audio, sample_audio, temp_dir, ffmpeg_available):
		"""Test muxing into MKV container."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		result = mux_audio_video.fn(
			video_path=sample_video_no_audio,
			audio_path=sample_audio,
			output_dir=str(temp_dir),
			label="test",
			container="mkv",
		)

		assert Path(result).exists()
		assert result.endswith(".mkv")

	def test_mux_creates_output_dir(self, sample_video_no_audio, sample_audio, temp_dir, ffmpeg_available):
		"""Test that mux creates the output directory if needed."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		nested_dir = temp_dir / "nested" / "final"

		result = mux_audio_video.fn(
			video_path=sample_video_no_audio,
			audio_path=sample_audio,
			output_dir=str(nested_dir),
			label="nested",
			container="mp4",
		)

		assert Path(result).exists()
		assert nested_dir.exists()
