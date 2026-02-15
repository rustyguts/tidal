import json
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from tidal.models.transcode import VMAFResult
from tidal.tasks.vmaf import calculate_vmaf, _build_vmaf_markdown


class TestCalculateVMAF:
	def test_vmaf_with_mock(self, sample_video, temp_dir):
		"""Test VMAF calculation with mocked ffmpeg call."""
		vmaf_data = {
			"pooled_metrics": {
				"vmaf": {
					"mean": 92.5,
					"min": 85.0,
					"max": 99.0,
					"harmonic_mean": 91.8,
				}
			}
		}

		mock_result = MagicMock()
		mock_result.returncode = 0
		mock_result.stderr = ""

		with (
			patch("subprocess.run", return_value=mock_result),
			patch("builtins.open", create=True) as mock_open,
			patch("json.load", return_value=vmaf_data),
			patch("tidal.tasks.vmaf.safe_create_progress", return_value=None),
			patch("tidal.tasks.vmaf.safe_update_progress"),
			patch("tidal.tasks.vmaf.create_markdown_artifact"),
		):
			# We need the file to exist for the check
			with patch.object(Path, "exists", return_value=True):
				with patch.object(Path, "unlink"):
					result = calculate_vmaf.fn(
						source_path=sample_video,
						encoded_path=sample_video,
						label="test",
					)

		assert result.score == 92.5
		assert result.min_score == 85.0
		assert result.max_score == 99.0
		assert result.harmonic_mean == 91.8
		assert result.quality_rating == "Good"

	def test_vmaf_nonexistent_source(self, tmp_path):
		"""Test that VMAF raises error for nonexistent source."""
		encoded = tmp_path / "encoded.mp4"
		encoded.write_bytes(b"fake")

		with pytest.raises(FileNotFoundError, match="source"):
			calculate_vmaf.fn(
				source_path="/nonexistent/source.mp4",
				encoded_path=str(encoded),
				label="test",
			)

	def test_vmaf_nonexistent_encoded(self, sample_video):
		"""Test that VMAF raises error for nonexistent encoded file."""
		with pytest.raises(FileNotFoundError, match="encoded"):
			calculate_vmaf.fn(
				source_path=sample_video,
				encoded_path="/nonexistent/encoded.mp4",
				label="test",
			)


class TestBuildVMAFMarkdown:
	def test_markdown_contains_score(self):
		result = VMAFResult(score=92.5, min_score=85.0, max_score=99.0, harmonic_mean=91.8)

		markdown = _build_vmaf_markdown(
			result=result,
			source_path="/data/source.mp4",
			encoded_path="/data/encoded.mp4",
			label="1080p",
		)

		assert "92.50" in markdown
		assert "85.00" in markdown
		assert "99.00" in markdown
		assert "Good" in markdown
		assert "1080p" in markdown
		assert "source.mp4" in markdown
		assert "encoded.mp4" in markdown

	def test_markdown_quality_ratings(self):
		"""Test that all quality ratings appear correctly in markdown."""
		for score, rating in [(97.0, "Excellent"), (85.0, "Good"), (65.0, "Fair"), (40.0, "Poor")]:
			result = VMAFResult(score=score, min_score=score - 10, max_score=score + 5, harmonic_mean=score - 1)
			markdown = _build_vmaf_markdown(result, "/src.mp4", "/enc.mp4", "test")
			assert rating in markdown
