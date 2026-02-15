import glob as glob_module
import subprocess
from pathlib import Path

from prefect import task

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.models.transcode import ProbeResult, SegmentResult
from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


@task(
	name="segment-video",
	description="Segment a video into chunks and extract audio",
	retries=2,
	retry_delay_seconds=5,
	tags=["segmentation"],
	task_run_name="segment-{source_path}",
)
def segment_video(
	source_path: str,
	work_dir: str,
	segment_duration: int = 10,
	probe: ProbeResult | None = None,
) -> SegmentResult:
	"""Segment a video file into video-only chunks and extract audio.

	Splits the source video into time-based segments using stream copy
	(no re-encoding) for speed. Audio is extracted to a separate file
	for independent processing.
	"""
	logger = get_logger("segment-video")
	src = Path(source_path)
	workdir = Path(work_dir)

	if not src.exists():
		raise FileNotFoundError(f"Source file does not exist: {source_path}")

	# Create working directories
	chunks_dir = workdir / "chunks"
	chunks_dir.mkdir(parents=True, exist_ok=True)

	logger.info(f"Segmenting {source_path} into {segment_duration}s chunks")

	# Progress artifact
	progress_id = safe_create_progress(
		0.0,
		f"Segmenting video: {src.name}",
	)

	# Segment video (video only, no audio)
	segment_pattern = str(chunks_dir / f"{src.stem}_%04d.mkv")

	processor = FFmpegProcessor()

	def on_progress(data: ProgressData) -> None:
		if data.progress_percent is not None:
			# Segmentation is roughly half the work (audio extraction is the other half)
			safe_update_progress(progress_id, data.progress_percent * 0.7)

	segment_args = [
		"-y",
		"-i",
		source_path,
		"-an",  # Strip audio
		"-c:v",
		"copy",  # Copy video codec (no re-encoding)
		"-f",
		"segment",
		"-segment_time",
		str(segment_duration),
		"-reset_timestamps",
		"1",
		segment_pattern,
	]

	processor.execute_sync(
		args=segment_args,
		input_file=source_path,
		progress_callback=on_progress,
	)

	# Gather chunk paths (sorted)
	chunk_paths = sorted(glob_module.glob(str(chunks_dir / f"{src.stem}_*.mkv")))

	if not chunk_paths:
		raise RuntimeError("Segmentation produced no chunks")

	logger.info(f"Segmentation produced {len(chunk_paths)} chunks")

	# Extract audio to a separate file (if source has audio)
	audio_path = None
	has_audio = probe.has_audio if probe else True

	if has_audio:
		audio_file = workdir / f"{src.stem}_audio.mkv"
		logger.info(f"Extracting audio to {audio_file}")

		try:
			audio_args = [
				"-y",
				"-i",
				source_path,
				"-vn",  # Strip video
				"-c:a",
				"copy",  # Copy audio codec
				str(audio_file),
			]

			processor.execute_sync(args=audio_args)
			audio_path = str(audio_file)
			logger.info("Audio extraction complete")
		except Exception as e:
			logger.warning(f"Audio extraction failed (source may have no audio): {e}")
			audio_path = None

	safe_update_progress(progress_id, 100.0)

	return SegmentResult(
		chunk_paths=chunk_paths,
		audio_path=audio_path,
		segment_count=len(chunk_paths),
		work_dir=work_dir,
	)
