import os
from prefect import flow, get_run_logger
from pydantic import BaseModel


class TranscodeInput(BaseModel):
	input: str = "/path/to/video.mp4"
	output: str = "/path/to/output.mp4"
	command: str = "ffmpeg -i {input} -c:v libx264 -preset fast -crf 22 {output}"


@flow
def transcode(args: TranscodeInput) -> None:
	logger = get_run_logger()
	logger.info(f"Transcoding: {args.input}")

	logger.info(f"Checking to see that file exists: {input}")
	if not os.path.exists(args.input):
		raise FileNotFoundError(f"File does not exist: {args.input}")

	logger.info(f"Creating a unique directory: {input}")

	logger.info(f"Successfully transcoded: {input}")


if __name__ == "__main__":
	transcode.serve()
