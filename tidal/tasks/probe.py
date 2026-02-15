import json
import subprocess
from pathlib import Path

from prefect import task

from tidal.utilities.logging import get_logger

from tidal.models.transcode import ProbeResult


@task(
	name="probe-video",
	description="Probe a video file with ffprobe to extract metadata",
	retries=2,
	retry_delay_seconds=5,
	task_run_name="probe-{source_path}",
)
def probe_video(source_path: str) -> ProbeResult:
	"""Probe a video file to extract metadata using ffprobe.

	Extracts duration, resolution, codecs, frame rate, and bitrate from
	the source video file. This information is used to plan the encoding
	pipeline.
	"""
	logger = get_logger("probe-video")
	path = Path(source_path)

	if not path.exists():
		raise FileNotFoundError(f"Source file does not exist: {source_path}")

	logger.info(f"Probing video file: {source_path}")

	result = subprocess.run(
		[
			"ffprobe",
			"-v",
			"quiet",
			"-print_format",
			"json",
			"-show_format",
			"-show_streams",
			source_path,
		],
		capture_output=True,
		text=True,
		timeout=30,
	)

	if result.returncode != 0:
		raise RuntimeError(f"ffprobe failed: {result.stderr}")

	data = json.loads(result.stdout)

	# Find video stream
	video_stream = None
	audio_stream = None
	for stream in data.get("streams", []):
		if stream["codec_type"] == "video" and video_stream is None:
			video_stream = stream
		elif stream["codec_type"] == "audio" and audio_stream is None:
			audio_stream = stream

	if video_stream is None:
		raise ValueError(f"No video stream found in {source_path}")

	# Parse frame rate (could be "30/1" or "29.97")
	fps_str = video_stream.get("r_frame_rate", "30/1")
	if "/" in fps_str:
		num, den = fps_str.split("/")
		frame_rate = float(num) / float(den) if float(den) != 0 else 30.0
	else:
		frame_rate = float(fps_str)

	# Parse bitrate
	format_info = data.get("format", {})
	bitrate = None
	if "bit_rate" in format_info:
		bitrate = int(format_info["bit_rate"])
	elif "bit_rate" in video_stream:
		bitrate = int(video_stream["bit_rate"])

	probe_result = ProbeResult(
		duration=float(format_info.get("duration", video_stream.get("duration", 0))),
		width=int(video_stream["width"]),
		height=int(video_stream["height"]),
		video_codec=video_stream["codec_name"],
		audio_codec=audio_stream["codec_name"] if audio_stream else None,
		frame_rate=round(frame_rate, 3),
		bitrate=bitrate,
		has_audio=audio_stream is not None,
	)

	logger.info(
		f"Probe complete: {probe_result.width}x{probe_result.height} "
		f"{probe_result.video_codec} {probe_result.duration:.1f}s "
		f"audio={probe_result.has_audio}"
	)

	return probe_result
