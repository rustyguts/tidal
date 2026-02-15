from pathlib import Path
from typing import Optional

from prefect import flow

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress
from pydantic import BaseModel

from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


class SimpleTranscodeInput(BaseModel):
	"""Input for a simple single-file transcode job."""

	source_path: str
	output_path: str
	video_codec: str = "libx264"
	audio_codec: str = "aac"
	preset: str = "medium"
	crf: int = 23


@flow(
	name="simple-transcode",
	description="Simple single-file video transcode using FFmpegProcessor",
	log_prints=True,
	flow_run_name="transcode-{input.source_path}",
)
def transcode(input: SimpleTranscodeInput) -> str:
	"""Simple single-file transcode flow.

	This is a utility flow for straightforward transcoding of a single
	file without chunked parallelism. Useful for quick one-off jobs
	or audio-only transcodes.
	"""
	logger = get_logger("simple-transcode")

	source = Path(input.source_path)
	if not source.exists():
		raise FileNotFoundError(f"Source file does not exist: {input.source_path}")

	output = Path(input.output_path)
	output.parent.mkdir(parents=True, exist_ok=True)

	logger.info(f"Transcoding: {source.name} -> {output.name}")

	progress_id = safe_create_progress(
		0.0,
		f"Transcoding {source.name}",
	)

	processor = FFmpegProcessor()

	def on_progress(data: ProgressData) -> None:
		if data.progress_percent is not None:
			safe_update_progress(progress_id, data.progress_percent)

	args = [
		"-y",
		"-i",
		input.source_path,
		"-c:v",
		input.video_codec,
		"-preset",
		input.preset,
		"-crf",
		str(input.crf),
		"-c:a",
		input.audio_codec,
		input.output_path,
	]

	processor.execute_sync(
		args=args,
		input_file=input.source_path,
		progress_callback=on_progress,
	)

	if not output.exists():
		raise RuntimeError(f"Transcoding failed: output not created at {output}")

	safe_update_progress(progress_id, 100.0)

	file_size_mb = output.stat().st_size / (1024 * 1024)
	logger.info(f"Transcode complete: {output.name} ({file_size_mb:.1f} MB)")

	return str(output)
