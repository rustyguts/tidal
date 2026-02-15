from pathlib import Path

from prefect import task

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


@task(
	name="mux-audio-video",
	description="Mux transcoded audio into the final video file",
	retries=2,
	retry_delay_seconds=5,
	task_run_name="mux-{label}",
)
def mux_audio_video(
	video_path: str,
	audio_path: str,
	output_dir: str,
	label: str = "output",
	container: str = "mp4",
) -> str:
	"""Mux audio and video streams into a single output file.

	Takes the concatenated video file and the transcoded audio file
	and combines them into the final output container (default: mp4).
	Uses stream copy to avoid re-encoding.
	"""
	logger = get_logger("mux-audio-video")
	out_dir = Path(output_dir)
	out_dir.mkdir(parents=True, exist_ok=True)

	output_file = out_dir / f"final_{label}.{container}"

	logger.info(f"Muxing audio + video -> {output_file.name}")

	progress_id = safe_create_progress(
		0.0,
		f"Muxing audio into video [{label}]",
	)

	processor = FFmpegProcessor()

	def on_progress(data: ProgressData) -> None:
		if data.progress_percent is not None:
			safe_update_progress(progress_id, data.progress_percent)

	args = [
		"-y",
		"-i",
		video_path,
		"-i",
		audio_path,
		"-c:v",
		"copy",  # Copy video stream
		"-c:a",
		"copy",  # Copy audio stream (already transcoded)
		"-shortest",  # Match duration to shortest stream
		str(output_file),
	]

	processor.execute_sync(
		args=args,
		input_file=video_path,
		progress_callback=on_progress,
	)

	if not output_file.exists():
		raise RuntimeError(f"Muxing failed: output not created at {output_file}")

	file_size_mb = output_file.stat().st_size / (1024 * 1024)
	logger.info(f"Muxing complete: {output_file.name} ({file_size_mb:.1f} MB)")

	safe_update_progress(progress_id, 100.0)

	return str(output_file)
