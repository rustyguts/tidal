import json
import subprocess
from unittest.mock import MagicMock, patch

import pytest

from tidal.tasks.probe import probe_video


class TestProbeVideo:
	def test_probe_real_video(self, sample_video, ffprobe_available):
		"""Test probing a real video file."""
		if not ffprobe_available:
			pytest.skip("ffprobe not available")

		result = probe_video.fn(source_path=sample_video)

		assert result.width == 320
		assert result.height == 240
		assert result.video_codec == "h264"
		assert result.has_audio is True
		assert result.audio_codec == "aac"
		assert result.duration > 0
		assert result.frame_rate > 0

	def test_probe_video_no_audio(self, sample_video_no_audio, ffprobe_available):
		"""Test probing a video file without audio."""
		if not ffprobe_available:
			pytest.skip("ffprobe not available")

		result = probe_video.fn(source_path=sample_video_no_audio)

		assert result.width == 320
		assert result.height == 240
		assert result.has_audio is False
		assert result.audio_codec is None

	def test_probe_nonexistent_file(self):
		"""Test that probing a nonexistent file raises FileNotFoundError."""
		with pytest.raises(FileNotFoundError, match="does not exist"):
			probe_video.fn(source_path="/nonexistent/video.mp4")

	def test_probe_parses_ffprobe_json(self, tmp_path, sample_probe_json):
		"""Test that probe correctly parses ffprobe JSON output."""
		# Create a dummy file so the path check passes
		dummy = tmp_path / "video.mp4"
		dummy.write_bytes(b"fake video data")

		mock_result = MagicMock()
		mock_result.returncode = 0
		mock_result.stdout = json.dumps(sample_probe_json)

		with patch("subprocess.run", return_value=mock_result):
			result = probe_video.fn(source_path=str(dummy))

		assert result.width == 1920
		assert result.height == 1080
		assert result.video_codec == "h264"
		assert result.audio_codec == "aac"
		assert result.duration == 120.5
		assert result.frame_rate == 30.0
		assert result.bitrate == 5500000
		assert result.has_audio is True

	def test_probe_handles_fractional_fps(self, tmp_path):
		"""Test parsing of fractional frame rates like 30000/1001."""
		dummy = tmp_path / "video.mp4"
		dummy.write_bytes(b"fake")

		probe_json = {
			"streams": [
				{
					"index": 0,
					"codec_type": "video",
					"codec_name": "h264",
					"width": 1920,
					"height": 1080,
					"r_frame_rate": "30000/1001",
				}
			],
			"format": {"duration": "10.0"},
		}

		mock_result = MagicMock()
		mock_result.returncode = 0
		mock_result.stdout = json.dumps(probe_json)

		with patch("subprocess.run", return_value=mock_result):
			result = probe_video.fn(source_path=str(dummy))

		assert abs(result.frame_rate - 29.97) < 0.01

	def test_probe_no_video_stream(self, tmp_path):
		"""Test that probe raises ValueError if no video stream found."""
		dummy = tmp_path / "audio_only.mp3"
		dummy.write_bytes(b"fake")

		probe_json = {
			"streams": [
				{
					"index": 0,
					"codec_type": "audio",
					"codec_name": "aac",
				}
			],
			"format": {"duration": "10.0"},
		}

		mock_result = MagicMock()
		mock_result.returncode = 0
		mock_result.stdout = json.dumps(probe_json)

		with patch("subprocess.run", return_value=mock_result):
			with pytest.raises(ValueError, match="No video stream"):
				probe_video.fn(source_path=str(dummy))

	def test_probe_ffprobe_failure(self, tmp_path):
		"""Test that probe raises RuntimeError on ffprobe failure."""
		dummy = tmp_path / "video.mp4"
		dummy.write_bytes(b"fake")

		mock_result = MagicMock()
		mock_result.returncode = 1
		mock_result.stderr = "ffprobe error"

		with patch("subprocess.run", return_value=mock_result):
			with pytest.raises(RuntimeError, match="ffprobe failed"):
				probe_video.fn(source_path=str(dummy))
