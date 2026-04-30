package ffmpeg

import (
	"os"
	"syscall"
)

func mkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func interruptSignal() os.Signal {
	return syscall.SIGINT
}
