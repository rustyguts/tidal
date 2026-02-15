import uuid
from pathlib import Path

from prefect import flow
from prefect.artifacts import create_markdown_artifact

from tidal.utilities.logging import get_logger

from tidal.flows.encode import encode_resolution
from tidal.models.transcode import CodecConfig, TranscodeJobInput, VideoResolution
from tidal.tasks.audio_transcode import transcode_audio
from tidal.tasks.mux import mux_audio_video
from tidal.tasks.probe import probe_video
from tidal.tasks.segmentation import segment_video
from tidal.tasks.vmaf import calculate_vmaf


@flow(
	name="tidal-pipeline",
	description="Main video processing pipeline: segment, encode in parallel, mux, and score",
	log_prints=True,
	flow_run_name="pipeline-{input.source_path}",
	retries=0,
)
def pipeline(input: TranscodeJobInput) -> dict:
	"""Main orchestrator flow for the Tidal video processing pipeline.

	This flow coordinates the full video processing pipeline:
	1. Probe the source video for metadata
	2. Segment the video into chunks (video-only) and extract audio
	3. For each target resolution, run a sub-flow to encode chunks in parallel
	4. Transcode audio independently
	5. Mux audio into each encoded video
	6. Calculate VMAF quality score on the primary output

	The pipeline is designed for parallelism:
	- Chunk encoding within a resolution runs in parallel via task submission
	- Multiple resolutions are processed sequentially (each is a sub-flow)
	- Audio transcoding runs as a submitted task alongside video encoding
	- VMAF scoring happens after muxing

	Args:
		input: TranscodeJobInput with source path and encoding parameters

	Returns:
		Dict with output file paths and VMAF results
	"""
	logger = get_logger("pipeline")

	source = Path(input.source_path)
	logger.info(f"Starting Tidal pipeline for: {source.name}")

	# Create unique working directory for this job
	job_id = uuid.uuid4().hex[:8]
	output_base = Path(input.output_dir) if input.output_dir else source.parent / "output"
	work_dir = output_base / f"tidal_{source.stem}_{job_id}"
	work_dir.mkdir(parents=True, exist_ok=True)

	logger.info(f"Working directory: {work_dir}")

	# ── Step 1: Probe source video ──────────────────────────────────────
	logger.info("Step 1/6: Probing source video...")
	probe = probe_video(source_path=input.source_path)

	logger.info(f"Source: {probe.width}x{probe.height} {probe.video_codec} {probe.duration:.1f}s audio={probe.has_audio}")

	# ── Step 2: Determine target resolutions ────────────────────────────
	if input.resolutions:
		target_resolutions = input.resolutions
	else:
		# Default: transcode at source resolution
		target_resolutions = [
			VideoResolution(
				width=probe.width,
				height=probe.height,
				label=f"{probe.height}p",
			)
		]

	logger.info(f"Target resolutions: {[r.label for r in target_resolutions]}")

	# ── Step 3: Segment video and extract audio ─────────────────────────
	logger.info("Step 2/6: Segmenting video...")
	segments = segment_video(
		source_path=input.source_path,
		work_dir=str(work_dir),
		segment_duration=input.segment_duration,
		probe=probe,
	)

	logger.info(f"Segmentation complete: {segments.segment_count} chunks")

	# ── Step 4: Start audio transcoding (submitted for concurrent execution) ──
	audio_future = None
	if segments.audio_path and probe.has_audio:
		logger.info("Step 3/6: Submitting audio transcode...")
		audio_future = transcode_audio.submit(
			audio_path=segments.audio_path,
			output_dir=str(work_dir),
			codec=input.codec,
		)
	else:
		logger.info("Step 3/6: No audio to transcode, skipping")

	# ── Step 5: Encode each resolution (sub-flows) ─────────────────────
	logger.info("Step 4/6: Encoding video chunks...")
	encoded_videos: dict[str, str] = {}

	for resolution in target_resolutions:
		res_output_dir = str(work_dir / f"encoded_{resolution.label}")

		# Determine whether we need to scale (only if different from source)
		scale_resolution = None
		if resolution.width != probe.width or resolution.height != probe.height:
			scale_resolution = resolution

		# Run encoding sub-flow (this handles parallel chunk encoding + concatenation)
		video_path = encode_resolution(
			chunk_paths=segments.chunk_paths,
			resolution_label=resolution.label,
			codec=input.codec,
			output_dir=res_output_dir,
			container=input.container,
			resolution=scale_resolution,
		)

		encoded_videos[resolution.label] = video_path
		logger.info(f"Resolution [{resolution.label}] encoding complete: {video_path}")

	# ── Step 6: Wait for audio and mux ─────────────────────────────────
	logger.info("Step 5/6: Muxing audio and video...")
	final_outputs: dict[str, str] = {}

	if audio_future:
		transcoded_audio = audio_future.result()

		for label, video_path in encoded_videos.items():
			final_path = mux_audio_video(
				video_path=video_path,
				audio_path=transcoded_audio,
				output_dir=str(work_dir / "final"),
				label=label,
				container=input.container,
			)
			final_outputs[label] = final_path
			logger.info(f"Muxed [{label}]: {final_path}")
	else:
		# No audio -- the encoded videos are already the final outputs
		final_outputs = encoded_videos
		logger.info("No audio to mux, encoded videos are final")

	# ── Step 7: Calculate VMAF on primary output ────────────────────────
	logger.info("Step 6/6: Calculating VMAF quality score...")
	primary_label = target_resolutions[0].label
	primary_output = final_outputs[primary_label]

	vmaf_result = calculate_vmaf(
		source_path=input.source_path,
		encoded_path=primary_output,
		label=primary_label,
	)

	# ── Create summary artifact ─────────────────────────────────────────
	_create_pipeline_summary(
		source_name=source.name,
		probe=probe,
		input=input,
		final_outputs=final_outputs,
		vmaf_score=vmaf_result.score,
		vmaf_rating=vmaf_result.quality_rating,
	)

	logger.info(
		f"Pipeline complete! VMAF={vmaf_result.score:.2f} ({vmaf_result.quality_rating}) "
		f"Outputs: {list(final_outputs.keys())}"
	)

	return {
		"outputs": final_outputs,
		"vmaf": {
			"score": vmaf_result.score,
			"rating": vmaf_result.quality_rating,
			"min": vmaf_result.min_score,
			"max": vmaf_result.max_score,
		},
		"work_dir": str(work_dir),
	}


def _create_pipeline_summary(
	source_name: str,
	probe,
	input: TranscodeJobInput,
	final_outputs: dict[str, str],
	vmaf_score: float,
	vmaf_rating: str,
) -> None:
	"""Create a summary markdown artifact for the pipeline run."""
	outputs_table = "\n".join(f"| {label} | `{Path(path).name}` |" for label, path in final_outputs.items())

	markdown = f"""# Tidal Pipeline Summary

## Source
- **File**: `{source_name}`
- **Resolution**: {probe.width}x{probe.height}
- **Duration**: {probe.duration:.1f}s
- **Video Codec**: {probe.video_codec}
- **Audio**: {"Yes (" + probe.audio_codec + ")" if probe.has_audio else "None"}

## Encoding Configuration
- **Video Codec**: {input.codec.video_codec}
- **Preset**: {input.codec.video_preset}
- **CRF**: {input.codec.crf}
- **Audio Codec**: {input.codec.audio_codec} @ {input.codec.audio_bitrate}
- **Container**: {input.container}

## Outputs

| Resolution | File |
|------------|------|
{outputs_table}

## Quality
- **VMAF Score**: **{vmaf_score:.2f}** ({vmaf_rating})
"""

	create_markdown_artifact(
		key="pipeline-summary",
		markdown=markdown,
		description=f"Tidal pipeline summary for {source_name}",
	)
