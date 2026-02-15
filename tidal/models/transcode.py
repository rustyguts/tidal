from __future__ import annotations

from pathlib import Path
from typing import Optional

from pydantic import BaseModel, field_validator


class VideoResolution(BaseModel):
	"""Represents a target video resolution."""

	width: int
	height: int
	label: str  # e.g. "1080p", "720p", "source"

	@field_validator("width", "height")
	@classmethod
	def must_be_positive_even(cls, v: int) -> int:
		if v <= 0:
			raise ValueError("Resolution dimensions must be positive")
		if v % 2 != 0:
			raise ValueError("Resolution dimensions must be even for FFmpeg compatibility")
		return v


class CodecConfig(BaseModel):
	"""Codec and encoding configuration."""

	video_codec: str = "libx264"
	video_preset: str = "medium"
	audio_codec: str = "aac"
	audio_bitrate: str = "192k"
	crf: int = 23
	pixel_format: str = "yuv420p"

	@field_validator("crf")
	@classmethod
	def crf_in_range(cls, v: int) -> int:
		if not 0 <= v <= 51:
			raise ValueError("CRF must be between 0 and 51")
		return v


class TranscodeJobInput(BaseModel):
	"""Input parameters for the main pipeline flow."""

	source_path: str
	output_dir: Optional[str] = None
	resolutions: Optional[list[VideoResolution]] = None
	codec: CodecConfig = CodecConfig()
	segment_duration: int = 10
	container: str = "mp4"

	@field_validator("source_path")
	@classmethod
	def source_must_exist(cls, v: str) -> str:
		if not Path(v).exists():
			raise ValueError(f"Source file does not exist: {v}")
		return v

	@field_validator("segment_duration")
	@classmethod
	def segment_duration_positive(cls, v: int) -> int:
		if v <= 0:
			raise ValueError("Segment duration must be positive")
		return v

	@field_validator("container")
	@classmethod
	def container_supported(cls, v: str) -> str:
		supported = {"mp4", "mkv", "webm", "mov"}
		if v not in supported:
			raise ValueError(f"Container '{v}' not supported. Use one of: {supported}")
		return v


class ProbeResult(BaseModel):
	"""Result from probing a video file with ffprobe."""

	duration: float
	width: int
	height: int
	video_codec: str
	audio_codec: Optional[str] = None
	frame_rate: float
	bitrate: Optional[int] = None
	has_audio: bool = True

	@property
	def resolution_label(self) -> str:
		return f"{self.height}p"


class SegmentResult(BaseModel):
	"""Result from segmenting a video file."""

	chunk_paths: list[str]
	audio_path: Optional[str] = None
	segment_count: int
	work_dir: str


class VMAFResult(BaseModel):
	"""Result from VMAF quality calculation."""

	score: float
	min_score: float
	max_score: float
	harmonic_mean: float

	@property
	def quality_rating(self) -> str:
		if self.score >= 95:
			return "Excellent"
		elif self.score >= 80:
			return "Good"
		elif self.score >= 60:
			return "Fair"
		else:
			return "Poor"
