import tempfile
from pathlib import Path

from prefect import task

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


@task(
	name="concatenate-chunks",
	description="Concatenate encoded video chunks into a single video file",
	retries=2,
	retry_delay_seconds=5,
	tags=["encoding"],
	task_run_name="concatenate-{label}",
)
def concatenate_chunks(
	chunk_paths: list[str],
	output_dir: str,
	label: str = "output",
	container: str = "mp4",
) -> str:
	"""Concatenate encoded video chunks back into a single video file.

	Uses FFmpeg's concat demuxer to join encoded chunks in order.
	This avoids re-encoding -- it simply joins the bitstreams. The
	intermediate container is MKV for compatibility; final muxing
	into the target container happens in the mux step.
	"""
	logger = get_logger("concatenate-chunks")
	out_dir = Path(output_dir)
	out_dir.mkdir(parents=True, exist_ok=True)

	output_file = out_dir / f"video_{label}.mkv"

	if not chunk_paths:
		raise ValueError("No chunks to concatenate")

	logger.info(f"Concatenating {len(chunk_paths)} chunks for [{label}]")

	progress_id = safe_create_progress(
		0.0,
		f"Concatenating {len(chunk_paths)} chunks [{label}]",
	)

	# Create concat list file
	concat_list = tempfile.NamedTemporaryFile(
		mode="w",
		suffix=".txt",
		dir=str(out_dir),
		delete=False,
		prefix="concat_",
	)

	try:
		for chunk_path in chunk_paths:
			# FFmpeg concat demuxer requires escaped single quotes in paths
			escaped_path = chunk_path.replace("'", "'\\''")
			concat_list.write(f"file '{escaped_path}'\n")
		concat_list.close()

		processor = FFmpegProcessor()

		def on_progress(data: ProgressData) -> None:
			if data.progress_percent is not None:
				safe_update_progress(progress_id, data.progress_percent)

		args = [
			"-y",
			"-f",
			"concat",
			"-safe",
			"0",
			"-i",
			concat_list.name,
			"-c",
			"copy",  # Stream copy (no re-encoding)
			str(output_file),
		]

		processor.execute_sync(
			args=args,
			progress_callback=on_progress,
		)

	finally:
		# Clean up concat list file
		try:
			Path(concat_list.name).unlink()
		except OSError:
			pass

	if not output_file.exists():
		raise RuntimeError(f"Concatenation failed: output not created at {output_file}")

	file_size_mb = output_file.stat().st_size / (1024 * 1024)
	logger.info(f"Concatenation complete: {output_file.name} ({file_size_mb:.1f} MB)")

	safe_update_progress(progress_id, 100.0)

	return str(output_file)
