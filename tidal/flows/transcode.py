import os
from prefect import flow, get_run_logger
from pydantic import BaseModel
from prefect.artifacts import (
    create_progress_artifact,
    update_progress_artifact,
)
import subprocess


class TranscodeInput(BaseModel):
	ffmpeg_arguments: str


@flow
def transcode(input: TranscodeInput) -> None:
	logger = get_run_logger()
	
	progress_artifact_id = create_progress_artifact(
		progress=0.0,
		description="Starting transcoding process",
	)

	result = subprocess.run(
		f"ffmpeg {input.ffmpeg_arguments}",
		text=True,
		shell=True,
		capture_output=True,
	)

	update_progress_artifact(100)

	if result.returncode != 0:
		logger.error(f"FFmpeg failed with return code {result.returncode}: {result.stderr}")
		raise RuntimeError(f"FFmpeg process failed: {result.stderr}")
	else:
		logger.info(f"FFmpeg completed successfully: {result.stdout}")

if __name__ == "__main__":
	transcode.serve()
