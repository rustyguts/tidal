package jobs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// moveFile relocates `src` to `dest`. If `dest` is a directory (or ends with /),
// the file is placed inside it under its original basename. Cross-mount
// fallback uses copy + delete. Creates the destination dir if missing.
func moveFile(src, dest string) error {
	if dest == "" {
		return errors.New("dest required")
	}
	target := dest
	if isDirHint(dest) {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return fmt.Errorf("mkdir dest: %w", err)
		}
		target = filepath.Join(dest, filepath.Base(src))
	} else {
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("mkdir parent: %w", err)
		}
	}

	if err := os.Rename(src, target); err == nil {
		return nil
	}
	// Cross-device fallback.
	if err := copyFile(src, target); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("remove source after copy: %w", err)
	}
	return nil
}

// isDirHint returns true when dest looks like a directory (trailing slash or
// existing directory). Files with no extension look like dirs only if the path
// already exists and is a directory.
func isDirHint(p string) bool {
	if len(p) > 0 && (p[len(p)-1] == '/' || p[len(p)-1] == os.PathSeparator) {
		return true
	}
	if fi, err := os.Stat(p); err == nil && fi.IsDir() {
		return true
	}
	return false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open dst: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return out.Sync()
}
