from prefect import flow, get_run_logger


@flow
def chunked_transcode(input: str) -> None:
	logger = get_run_logger()
	logger.info(f"Transcoding: {input}")
	
	logger.info(f"Checking to see that file exists: {input}")

	logger.info(f"Creating a unique directory: {input}")
	
	logger.info(f"Splitting file into chunks: {input}")

	logger.info(f"Enqueuing transcode jobs: {input}")

	logger.info(f"Concatenating video chunks: {input}")

	logger.info(f"Muxing transcoded video with audio: {input}")

	logger.info(f"Successfully transcoded: {input}")


if __name__ == "__main__":
	chunked_transcode.serve()
