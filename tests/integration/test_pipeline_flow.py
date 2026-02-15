from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from tidal.flows.pipeline import pipeline
from tidal.models.transcode import CodecConfig, TranscodeJobInput, VideoResolution


@pytest.mark.integration
class TestPipelineFlow:
	def test_pipeline_source_resolution(self, sample_video, temp_dir, ffmpeg_available):
		"""Test full pipeline at source resolution.

		VMAF is mocked since it requires libvmaf in the FFmpeg build.
		"""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		output_dir = str(temp_dir / "output")

		mock_vmaf_result = MagicMock()
		mock_vmaf_result.score = 95.0
		mock_vmaf_result.min_score = 90.0
		mock_vmaf_result.max_score = 99.0
		mock_vmaf_result.harmonic_mean = 94.5
		mock_vmaf_result.quality_rating = "Excellent"

		input_model = TranscodeJobInput(
			source_path=sample_video,
			output_dir=output_dir,
			codec=CodecConfig(video_codec="libx264", preset="ultrafast", crf=28),
			segment_duration=1,
		)

		with patch("tidal.flows.pipeline.calculate_vmaf", return_value=mock_vmaf_result):
			with patch("tidal.flows.pipeline.create_markdown_artifact"):
				# Call the flow directly (not .fn()) so Prefect sets up the full context
				result = pipeline(input=input_model)

		assert "outputs" in result
		assert "vmaf" in result
		assert result["vmaf"]["score"] == 95.0

		for label, path in result["outputs"].items():
			assert Path(path).exists(), f"Output for {label} not found at {path}"

	def test_pipeline_custom_resolution(self, sample_video, temp_dir, ffmpeg_available):
		"""Test pipeline with a custom target resolution."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		output_dir = str(temp_dir / "output")

		mock_vmaf_result = MagicMock()
		mock_vmaf_result.score = 88.0
		mock_vmaf_result.min_score = 80.0
		mock_vmaf_result.max_score = 95.0
		mock_vmaf_result.harmonic_mean = 87.0
		mock_vmaf_result.quality_rating = "Good"

		input_model = TranscodeJobInput(
			source_path=sample_video,
			output_dir=output_dir,
			resolutions=[VideoResolution(width=160, height=120, label="120p")],
			codec=CodecConfig(video_codec="libx264", preset="ultrafast", crf=28),
			segment_duration=1,
		)

		with patch("tidal.flows.pipeline.calculate_vmaf", return_value=mock_vmaf_result):
			with patch("tidal.flows.pipeline.create_markdown_artifact"):
				result = pipeline(input=input_model)

		assert "120p" in result["outputs"]
		assert Path(result["outputs"]["120p"]).exists()
