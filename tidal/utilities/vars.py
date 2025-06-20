from pydantic import BaseModel


class GlobalQueueConfig(BaseModel):
	name: str
	limit: int
	slot_decay_per_second: float


class TaskQueueConfig(BaseModel):
	name: str
	limit: int


TEST_GLOBAL_QUEUE = GlobalQueueConfig(
	name="test",
	limit=4,
	slot_decay_per_second=1,
)

TEST_TASK_QUEUE = TaskQueueConfig(
	name="test",
	limit=3,
)

TaskQueues = [TEST_TASK_QUEUE]
GlobalQueues = [TEST_GLOBAL_QUEUE]
