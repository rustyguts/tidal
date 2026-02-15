from pathlib import Path

from prefect import task

from tidal.utilities.logging import get_logger, safe_create_progress, safe_update_progress

from tidal.models.transcode import CodecConfig
from tidal.utilities.ffmpeg import FFmpegProcessor, ProgressData


@task(
	name="transcode-audio",
	description="Transcode the audio stream to the target codec",
	retries=2,
	retry_delay_seconds=5,
	tags=["encoding"],
	task_run_name="transcode-audio-{audio_codec}",
)
def transcode_audio(
	audio_path: str,
	output_dir: str,
	codec: CodecConfig,
	audio_codec: str = "",
) -> str:
	"""Transcode audio to the target codec and bitrate.

	Takes the extracted audio file and transcodes it to the configured
	audio codec (default: AAC) and bitrate. Returns the path to the
	transcoded audio file.
	"""
	logger = get_logger("transcode-audio")
	src = Path(audio_path)
	out_dir = Path(output_dir)
	out_dir.mkdir(parents=True, exist_ok=True)

	effective_codec = audio_codec or codec.audio_codec
	output_file = out_dir / f"audio_transcoded.{_codec_extension(effective_codec)}"

	logger.info(f"Transcoding audio: {effective_codec} @ {codec.audio_bitrate}")

	progress_id = safe_create_progress(
		0.0,
		f"Transcoding audio to {effective_codec}",
	)

	processor = FFmpegProcessor()

	def on_progress(data: ProgressData) -> None:
		if data.progress_percent is not None:
			safe_update_progress(progress_id, data.progress_percent)

	args = [
		"-y",
		"-i",
		audio_path,
		"-vn",  # No video
		"-c:a",
		effective_codec,
		"-b:a",
		codec.audio_bitrate,
		str(output_file),
	]

	processor.execute_sync(
		args=args,
		input_file=audio_path,
		progress_callback=on_progress,
	)

	if not output_file.exists():
		raise RuntimeError(f"Audio transcoding failed: output not created at {output_file}")

	file_size_mb = output_file.stat().st_size / (1024 * 1024)
	logger.info(f"Audio transcoded: {output_file.name} ({file_size_mb:.1f} MB)")

	safe_update_progress(progress_id, 100.0)

	return str(output_file)


def _codec_extension(codec: str) -> str:
	"""Map audio codec name to a suitable file extension."""
	mapping = {
		"aac": "m4a",
		"libopus": "opus",
		"opus": "opus",
		"libvorbis": "ogg",
		"flac": "flac",
		"libmp3lame": "mp3",
		"mp3": "mp3",
		"pcm_s16le": "wav",
	}
	return mapping.get(codec, "mka")
