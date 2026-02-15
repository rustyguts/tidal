import pytest
from pydantic import ValidationError

from tidal.models.transcode import (
	CodecConfig,
	ProbeResult,
	SegmentResult,
	TranscodeJobInput,
	VideoResolution,
	VMAFResult,
)


class TestVideoResolution:
	def test_valid_resolution(self):
		res = VideoResolution(width=1920, height=1080, label="1080p")
		assert res.width == 1920
		assert res.height == 1080
		assert res.label == "1080p"

	def test_720p_resolution(self):
		res = VideoResolution(width=1280, height=720, label="720p")
		assert res.width == 1280
		assert res.height == 720

	def test_negative_width_rejected(self):
		with pytest.raises(ValidationError, match="positive"):
			VideoResolution(width=-1920, height=1080, label="bad")

	def test_zero_height_rejected(self):
		with pytest.raises(ValidationError, match="positive"):
			VideoResolution(width=1920, height=0, label="bad")

	def test_odd_dimensions_rejected(self):
		with pytest.raises(ValidationError, match="even"):
			VideoResolution(width=1921, height=1080, label="bad")

	def test_odd_height_rejected(self):
		with pytest.raises(ValidationError, match="even"):
			VideoResolution(width=1920, height=1081, label="bad")


class TestCodecConfig:
	def test_defaults(self):
		config = CodecConfig()
		assert config.video_codec == "libx264"
		assert config.video_preset == "medium"
		assert config.audio_codec == "aac"
		assert config.audio_bitrate == "192k"
		assert config.crf == 23
		assert config.pixel_format == "yuv420p"

	def test_custom_codec(self):
		config = CodecConfig(video_codec="libx265", crf=28, video_preset="slow")
		assert config.video_codec == "libx265"
		assert config.crf == 28
		assert config.video_preset == "slow"

	def test_crf_too_high(self):
		with pytest.raises(ValidationError, match="51"):
			CodecConfig(crf=52)

	def test_crf_negative(self):
		with pytest.raises(ValidationError, match="0"):
			CodecConfig(crf=-1)

	def test_crf_boundary_low(self):
		config = CodecConfig(crf=0)
		assert config.crf == 0

	def test_crf_boundary_high(self):
		config = CodecConfig(crf=51)
		assert config.crf == 51


class TestTranscodeJobInput:
	def test_valid_input(self, sample_video):
		input_model = TranscodeJobInput(source_path=sample_video)
		assert input_model.source_path == sample_video
		assert input_model.output_dir is None
		assert input_model.resolutions is None
		assert input_model.segment_duration == 10
		assert input_model.container == "mp4"

	def test_nonexistent_source_rejected(self):
		with pytest.raises(ValidationError, match="does not exist"):
			TranscodeJobInput(source_path="/nonexistent/video.mp4")

	def test_custom_output_dir(self, sample_video):
		input_model = TranscodeJobInput(source_path=sample_video, output_dir="/tmp/output")
		assert input_model.output_dir == "/tmp/output"

	def test_custom_resolutions(self, sample_video):
		resolutions = [
			VideoResolution(width=1920, height=1080, label="1080p"),
			VideoResolution(width=1280, height=720, label="720p"),
		]
		input_model = TranscodeJobInput(source_path=sample_video, resolutions=resolutions)
		assert len(input_model.resolutions) == 2

	def test_custom_codec(self, sample_video):
		codec = CodecConfig(video_codec="libx265", crf=28)
		input_model = TranscodeJobInput(source_path=sample_video, codec=codec)
		assert input_model.codec.video_codec == "libx265"
		assert input_model.codec.crf == 28

	def test_invalid_segment_duration(self, sample_video):
		with pytest.raises(ValidationError, match="positive"):
			TranscodeJobInput(source_path=sample_video, segment_duration=0)

	def test_negative_segment_duration(self, sample_video):
		with pytest.raises(ValidationError, match="positive"):
			TranscodeJobInput(source_path=sample_video, segment_duration=-5)

	def test_unsupported_container(self, sample_video):
		with pytest.raises(ValidationError, match="not supported"):
			TranscodeJobInput(source_path=sample_video, container="avi")

	def test_supported_containers(self, sample_video):
		for container in ["mp4", "mkv", "webm", "mov"]:
			input_model = TranscodeJobInput(source_path=sample_video, container=container)
			assert input_model.container == container


class TestProbeResult:
	def test_basic_probe(self):
		probe = ProbeResult(
			duration=120.5,
			width=1920,
			height=1080,
			video_codec="h264",
			audio_codec="aac",
			frame_rate=30.0,
			bitrate=5000000,
		)
		assert probe.duration == 120.5
		assert probe.width == 1920
		assert probe.height == 1080
		assert probe.has_audio is True

	def test_no_audio(self):
		probe = ProbeResult(
			duration=60.0,
			width=1280,
			height=720,
			video_codec="h264",
			frame_rate=24.0,
			has_audio=False,
		)
		assert probe.has_audio is False
		assert probe.audio_codec is None

	def test_resolution_label(self):
		probe = ProbeResult(
			duration=10.0,
			width=1920,
			height=1080,
			video_codec="h264",
			frame_rate=30.0,
		)
		assert probe.resolution_label == "1080p"

	def test_resolution_label_720p(self):
		probe = ProbeResult(
			duration=10.0,
			width=1280,
			height=720,
			video_codec="h264",
			frame_rate=30.0,
		)
		assert probe.resolution_label == "720p"


class TestSegmentResult:
	def test_basic_segment_result(self):
		result = SegmentResult(
			chunk_paths=["/tmp/chunk_0000.mkv", "/tmp/chunk_0001.mkv"],
			audio_path="/tmp/audio.mkv",
			segment_count=2,
			work_dir="/tmp/work",
		)
		assert len(result.chunk_paths) == 2
		assert result.audio_path is not None
		assert result.segment_count == 2

	def test_no_audio_segment_result(self):
		result = SegmentResult(
			chunk_paths=["/tmp/chunk_0000.mkv"],
			segment_count=1,
			work_dir="/tmp/work",
		)
		assert result.audio_path is None


class TestVMAFResult:
	def test_excellent_quality(self):
		result = VMAFResult(score=97.5, min_score=92.0, max_score=99.8, harmonic_mean=96.8)
		assert result.quality_rating == "Excellent"

	def test_good_quality(self):
		result = VMAFResult(score=85.0, min_score=70.0, max_score=95.0, harmonic_mean=83.5)
		assert result.quality_rating == "Good"

	def test_fair_quality(self):
		result = VMAFResult(score=65.0, min_score=50.0, max_score=80.0, harmonic_mean=63.0)
		assert result.quality_rating == "Fair"

	def test_poor_quality(self):
		result = VMAFResult(score=45.0, min_score=20.0, max_score=60.0, harmonic_mean=40.0)
		assert result.quality_rating == "Poor"

	def test_boundary_excellent(self):
		result = VMAFResult(score=95.0, min_score=90.0, max_score=100.0, harmonic_mean=94.0)
		assert result.quality_rating == "Excellent"

	def test_boundary_good(self):
		result = VMAFResult(score=80.0, min_score=70.0, max_score=90.0, harmonic_mean=79.0)
		assert result.quality_rating == "Good"

	def test_boundary_fair(self):
		result = VMAFResult(score=60.0, min_score=50.0, max_score=70.0, harmonic_mean=59.0)
		assert result.quality_rating == "Fair"
