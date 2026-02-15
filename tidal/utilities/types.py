import enum


class Environment(str, enum.Enum):
	LOCAL = "local"
	STAGING = "staging"
	PRODUCTION = "production"
