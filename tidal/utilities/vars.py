from pydantic import BaseModel


class GlobalQueueConfig(BaseModel):
	name: str
	limit: int
	slot_decay_per_second: float


class TaskQueueConfig(BaseModel):
	name: str
	limit: int


# Global concurrency limits
ENCODING_GLOBAL_QUEUE = GlobalQueueConfig(
	name="encoding",
	limit=8,
	slot_decay_per_second=1.0,
)

# Task-level concurrency limits
ENCODING_TASK_QUEUE = TaskQueueConfig(
	name="encoding",
	limit=6,
)

SEGMENTATION_TASK_QUEUE = TaskQueueConfig(
	name="segmentation",
	limit=2,
)

VMAF_TASK_QUEUE = TaskQueueConfig(
	name="vmaf",
	limit=2,
)

TaskQueues = [ENCODING_TASK_QUEUE, SEGMENTATION_TASK_QUEUE, VMAF_TASK_QUEUE]
GlobalQueues = [ENCODING_GLOBAL_QUEUE]
