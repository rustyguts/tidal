import json
import subprocess
import tempfile
from pathlib import Path

from prefect import task
from prefect.artifacts import create_markdown_artifact

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.models.transcode import VMAFResult


@task(
	name="calculate-vmaf",
	description="Calculate VMAF quality score between source and encoded video",
	retries=1,
	retry_delay_seconds=10,
	tags=["vmaf"],
	task_run_name="vmaf-{label}",
)
def calculate_vmaf(
	source_path: str,
	encoded_path: str,
	label: str = "output",
) -> VMAFResult:
	"""Calculate VMAF quality score comparing source to encoded video.

	Uses FFmpeg's libvmaf filter to compute the Video Multi-Method
	Assessment Fusion score. The result is stored as a Prefect markdown
	artifact for visibility in the UI.

	VMAF scores:
	  - 95+: Excellent (visually indistinguishable from source)
	  - 80-95: Good (minor artifacts, good for streaming)
	  - 60-80: Fair (noticeable quality loss)
	  - <60: Poor (significant quality degradation)
	"""
	logger = get_logger("calculate-vmaf")

	for path, name in [(source_path, "source"), (encoded_path, "encoded")]:
		if not Path(path).exists():
			raise FileNotFoundError(f"{name} file does not exist: {path}")

	logger.info(f"Calculating VMAF: {Path(encoded_path).name} vs {Path(source_path).name}")

	progress_id = safe_create_progress(
		0.0,
		f"Calculating VMAF score [{label}]",
	)

	# Create temp file for VMAF JSON output
	vmaf_log = tempfile.NamedTemporaryFile(
		suffix=".json",
		delete=False,
		prefix="vmaf_",
	)
	vmaf_log.close()

	try:
		# VMAF filter: distorted (encoded) is first input, reference (source) is second
		# We need to scale both to the same resolution for VMAF comparison
		vmaf_filter = (
			f"[0:v]setpts=PTS-STARTPTS[dist];"
			f"[1:v]setpts=PTS-STARTPTS[ref];"
			f"[dist][ref]libvmaf=log_path={vmaf_log.name}:log_fmt=json:n_threads=4"
		)

		cmd = [
			"ffmpeg",
			"-y",
			"-i",
			encoded_path,  # Distorted (encoded) video
			"-i",
			source_path,  # Reference (source) video
			"-lavfi",
			vmaf_filter,
			"-f",
			"null",
			"-",
		]

		logger.info(f"Running VMAF calculation...")
		safe_update_progress(progress_id, 10.0)

		result = subprocess.run(
			cmd,
			capture_output=True,
			text=True,
			timeout=3600,  # VMAF can take a while on long videos
		)

		if result.returncode != 0:
			raise RuntimeError(f"VMAF calculation failed: {result.stderr[-500:]}")

		safe_update_progress(progress_id, 80.0)

		# Parse VMAF results
		with open(vmaf_log.name) as f:
			vmaf_data = json.load(f)

		pooled = vmaf_data.get("pooled_metrics", {}).get("vmaf", {})

		vmaf_result = VMAFResult(
			score=pooled.get("mean", 0.0),
			min_score=pooled.get("min", 0.0),
			max_score=pooled.get("max", 0.0),
			harmonic_mean=pooled.get("harmonic_mean", 0.0),
		)

		logger.info(f"VMAF Score: {vmaf_result.score:.2f} ({vmaf_result.quality_rating})")

		# Create Prefect markdown artifact with VMAF results
		markdown = _build_vmaf_markdown(vmaf_result, source_path, encoded_path, label)
		create_markdown_artifact(
			key=f"vmaf-{label}",
			markdown=markdown,
			description=f"VMAF quality score for {label}",
		)

		safe_update_progress(progress_id, 100.0)

		return vmaf_result

	finally:
		try:
			Path(vmaf_log.name).unlink()
		except OSError:
			pass


def _build_vmaf_markdown(
	result: VMAFResult,
	source_path: str,
	encoded_path: str,
	label: str,
) -> str:
	"""Build a markdown summary of VMAF results for the Prefect artifact."""
	return f"""# VMAF Quality Report: {label}

## Score Summary

| Metric | Value |
|--------|-------|
| **Mean VMAF** | **{result.score:.2f}** |
| Min VMAF | {result.min_score:.2f} |
| Max VMAF | {result.max_score:.2f} |
| Harmonic Mean | {result.harmonic_mean:.2f} |
| **Quality Rating** | **{result.quality_rating}** |

## Files

- **Source**: `{Path(source_path).name}`
- **Encoded**: `{Path(encoded_path).name}`

## Quality Scale

| Range | Rating |
|-------|--------|
| 95+ | Excellent (visually indistinguishable) |
| 80-95 | Good (minor artifacts) |
| 60-80 | Fair (noticeable quality loss) |
| <60 | Poor (significant degradation) |
"""
