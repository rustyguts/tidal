from prefect import flow

from tidal.utilities.logging import get_logger
from prefect.futures import wait

from tidal.models.transcode import CodecConfig, VideoResolution
from tidal.tasks.concatenation import concatenate_chunks
from tidal.tasks.encode_chunk import encode_chunk


@flow(
	name="encode-resolution",
	description="Encode all video chunks for a single resolution, then concatenate",
	log_prints=True,
	flow_run_name="encode-{resolution_label}",
)
def encode_resolution(
	chunk_paths: list[str],
	resolution_label: str,
	codec: CodecConfig,
	output_dir: str,
	container: str = "mp4",
	resolution: VideoResolution | None = None,
) -> str:
	"""Sub-flow: encode all chunks for a single target resolution.

	Distributes chunk encoding across workers via .submit() for parallel
	execution, waits for all chunks to complete, then concatenates the
	encoded chunks into a single video file.

	This flow is designed to be called as a child flow from the main
	pipeline, once per target resolution. In production with a Kubernetes
	work pool, each chunk encoding task can run on a separate worker.

	Args:
		chunk_paths: List of source video chunk file paths
		resolution_label: Label for this resolution (e.g. "1080p", "source")
		codec: Codec configuration for encoding
		output_dir: Directory to write encoded chunks and concatenated output
		container: Output container format
		resolution: Target resolution (None = encode at source resolution)

	Returns:
		Path to the concatenated video file for this resolution
	"""
	logger = get_logger("encode")

	logger.info(
		f"Starting encoding for [{resolution_label}]: {len(chunk_paths)} chunks, codec={codec.video_codec}, crf={codec.crf}"
	)

	# Submit all chunk encoding tasks in parallel
	futures = []
	for i, chunk_path in enumerate(chunk_paths):
		future = encode_chunk.submit(
			chunk_path=chunk_path,
			output_dir=output_dir,
			codec=codec,
			chunk_index=i,
			resolution=resolution,
			resolution_label=resolution_label,
		)
		futures.append(future)

	# Wait for all encoding futures to complete
	wait(futures)

	# Collect results (will raise if any task failed)
	encoded_paths = [f.result() for f in futures]

	logger.info(f"All {len(encoded_paths)} chunks encoded for [{resolution_label}], concatenating...")

	# Concatenate encoded chunks into a single video
	concatenated_path = concatenate_chunks(
		chunk_paths=encoded_paths,
		output_dir=output_dir,
		label=resolution_label,
		container=container,
	)

	logger.info(f"Encoding complete for [{resolution_label}]: {concatenated_path}")

	return concatenated_path
