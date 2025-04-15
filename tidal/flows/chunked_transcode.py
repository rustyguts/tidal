import os
from prefect import flow, get_run_logger
from pydantic import BaseModel

class ChunkedTranscodeInput(BaseModel):
	src: str = "/path/to/video.mp4"

@flow
def chunked_transcode(input: ChunkedTranscodeInput) -> None:
	logger = get_run_logger()
	logger.info(f"Transcoding: {input.src}")
	
	logger.info(f"Checking to see that file exists: {input}")
	if not os.path.exists(input.src):
		raise FileNotFoundError(f"File does not exist: {input.src}")

	logger.info(f"Creating a unique directory: {input}")
	
	logger.info(f"Splitting file into chunks: {input}")

	logger.info(f"Enqueuing transcode jobs: {input}")

	logger.info(f"Concatenating video chunks: {input}")

	logger.info(f"Muxing transcoded video with audio: {input}")

	logger.info(f"Successfully transcoded: {input}")


if __name__ == "__main__":
	chunked_transcode.serve()
