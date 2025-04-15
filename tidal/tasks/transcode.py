from prefect import task
import subprocess


@task
def transcode(input: str) -> None:
	subprocess.run(['ffmpeg', '-i', 'input.mp4', 'output.mp3'])