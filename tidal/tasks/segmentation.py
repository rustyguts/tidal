from pathlib import Path
from prefect import get_run_logger, task
import subprocess

from pydantic import BaseModel

class SegmentationInput(BaseModel):
	src: Path
	segment_time: int = 10  # in seconds

@task
def segmentation(input: SegmentationInput) -> None:
	logger = get_run_logger()

	logger.info(f"Segmentation: {input.src}")
	src_directory = Path(input.src).parent
	src_segments_dir = src_directory / "source_segments"
	src_segments_dir.mkdir(parents=True, exist_ok=True)
	segment_pattern = src_segments_dir / f"{input.src.stem}_%04d.mkv"

	subprocess.run([
			'ffmpeg',
			'-hide_banner',
			'-loglevel', 'error',
			'-i', str(input.src),
			'-an',
			'-c:v', 'copy',                  # Copy video codec (no re-encoding)
			'-f', 'segment',                 # Use segment mode
			'-segment_time', str(input.segment_time),  # Segment duration
			'-reset_timestamps', '1',        # Reset timestamps
			str(segment_pattern)             # Output pattern
	])
    