import logging
from typing import Optional
from uuid import UUID, uuid4

from prefect.exceptions import MissingContextError


def get_logger(name: str = "tidal") -> logging.Logger:
	"""Get a logger that works both inside and outside Prefect context.

	When called within a Prefect flow or task run, returns the Prefect
	run logger which provides structured logging integrated with the
	Prefect UI. When called outside Prefect context (e.g. in tests),
	falls back to a standard Python logger.
	"""
	try:
		from prefect import get_run_logger

		return get_run_logger()
	except MissingContextError:
		return logging.getLogger(name)


def safe_create_progress(progress: float, description: str) -> Optional[UUID]:
	"""Create a progress artifact, returning None if outside Prefect context."""
	try:
		from prefect.artifacts import create_progress_artifact

		return create_progress_artifact(progress=progress, description=description)
	except (MissingContextError, Exception):
		return None


def safe_update_progress(artifact_id: Optional[UUID], progress: float) -> None:
	"""Update a progress artifact, silently no-op if outside Prefect context."""
	if artifact_id is None:
		return
	try:
		from prefect.artifacts import update_progress_artifact

		update_progress_artifact(artifact_id=artifact_id, progress=progress)
	except (MissingContextError, Exception):
		pass
