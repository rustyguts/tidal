from pathlib import Path
from typing import Optional

from prefect import task

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.models.transcode import CodecConfig, VideoResolution
from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


@task(
	name="encode-chunk",
	description="Encode a single video chunk with the specified codec and resolution",
	retries=2,
	retry_delay_seconds=10,
	tags=["encoding"],
	task_run_name="encode-chunk-{chunk_index:04d}-{resolution_label}",
)
def encode_chunk(
	chunk_path: str,
	output_dir: str,
	codec: CodecConfig,
	chunk_index: int,
	resolution: Optional[VideoResolution] = None,
	resolution_label: str = "source",
) -> str:
	"""Encode a single video chunk.

	Takes a video-only chunk and encodes it with the specified codec,
	preset, CRF, and optionally scales to a target resolution. Returns
	the path to the encoded chunk.
	"""
	logger = get_logger("encode-chunk")
	src = Path(chunk_path)
	out_dir = Path(output_dir)
	out_dir.mkdir(parents=True, exist_ok=True)

	label = resolution.label if resolution else resolution_label
	output_file = out_dir / f"encoded_{label}_{chunk_index:04d}.mkv"

	logger.info(f"Encoding chunk {chunk_index} -> {label} ({codec.video_codec} crf={codec.crf})")

	progress_id = safe_create_progress(
		0.0,
		f"Encoding chunk {chunk_index:04d} [{label}]",
	)

	processor = FFmpegProcessor()

	def on_progress(data: ProgressData) -> None:
		if data.progress_percent is not None:
			safe_update_progress(progress_id, data.progress_percent)

	# Build FFmpeg arguments
	args = [
		"-y",
		"-i",
		chunk_path,
		"-c:v",
		codec.video_codec,
		"-preset",
		codec.video_preset,
		"-crf",
		str(codec.crf),
		"-pix_fmt",
		codec.pixel_format,
	]

	# Add scaling filter if resolution is specified
	if resolution:
		args.extend(
			[
				"-vf",
				f"scale={resolution.width}:{resolution.height}",
			]
		)

	# No audio in chunks
	args.extend(["-an", str(output_file)])

	processor.execute_sync(
		args=args,
		input_file=chunk_path,
		progress_callback=on_progress,
	)

	if not output_file.exists():
		raise RuntimeError(f"Encoding failed: output file not created at {output_file}")

	file_size_mb = output_file.stat().st_size / (1024 * 1024)
	logger.info(f"Chunk {chunk_index} encoded: {output_file.name} ({file_size_mb:.1f} MB)")

	safe_update_progress(progress_id, 100.0)

	return str(output_file)
