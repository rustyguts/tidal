# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Tidal is a parallel video transcoding pipeline orchestrated by Prefect 3. FFmpeg (BtbN GPL build with libvmaf) is installed in the Docker image; tasks shell out to `ffmpeg`/`ffprobe` on PATH.

## Commands

Dependency + venv via `uv` (lockfile authoritative).

```bash
uv sync --extra dev                                  # install runtime + dev deps
uv run ruff check .                                  # lint
uv run ruff format .                                 # format (tabs, double quotes, line-length 120)
uv run mypy tidal                                    # type check (pydantic plugin)
uv run pytest                                        # all tests (asyncio_mode=auto)
uv run pytest -m "not integration"                   # unit only
uv run pytest -m integration                         # integration only
uv run pytest tests/unit/test_vmaf.py::test_name     # single test
uv run coverage run -m pytest && uv run coverage report   # coverage (fail_under=70)
```

Local stack (Prefect server + worker, hot reload via `watchfiles`):

```bash
docker-compose up                                    # server :4200, worker rebuilds on tidal/ changes
```

Register/serve deployments against a running Prefect server:

```bash
PREFECT_API_URL=http://localhost:4200/api uv run deploy.py all --environment local
# subcommands: prefect_variables, prefect_concurrency_limits, prefect_flows, all
```

Production deploy to k3s (`cirus` work pool, namespace `prefect`, image `ghcr.io/rustyguts/tidal`): see [DEPLOY.md](DEPLOY.md). CI ([.github/workflows/publish.yml](.github/workflows/publish.yml)) builds + pushes a `sha-`tagged image on push to `main`/`dev` and on releases.

## Architecture

Two registered Prefect deployments at the top:

- `pipeline/tidal-pipeline` ([tidal/flows/pipeline.py](tidal/flows/pipeline.py)) — chunked parallel pipeline.
- `transcode/simple-transcode` ([tidal/flows/transcode.py](tidal/flows/transcode.py)) — single-file transcode, no chunking.

Pipeline DAG (worth understanding before editing tasks):

1. `probe_video` — ffprobe metadata into `ProbeResult`.
2. `segment_video` — split source into N video-only chunks + extract audio track.
3. `transcode_audio.submit(...)` — fires concurrently with video encoding.
4. Per `VideoResolution`, call sub-flow `encode-resolution` ([tidal/flows/encode.py](tidal/flows/encode.py)). Sub-flow `encode_chunk.submit()`s every chunk in parallel, `wait()`s, then `concatenate_chunks` builds one file per resolution.
5. `mux_audio_video` joins transcoded audio onto each per-resolution output.
6. `calculate_vmaf` scores the primary (first) resolution; flow emits a markdown artifact summary.

Parallelism boundaries: chunks within a resolution run in parallel; resolutions run sequentially; audio runs concurrently with video. On the k3s `cirus` work pool, each `.submit()`ed task can land on a separate Kubernetes job.

### Concurrency limits

[tidal/utilities/vars.py](tidal/utilities/vars.py) declares `TaskQueues` (tag-based per-task: `encoding=6`, `segmentation=2`, `vmaf=2`) and `GlobalQueues` (`encoding=8`, with `slot_decay_per_second`). `deploy.py prefect_concurrency_limits` reconciles these on the server. Tasks must carry the matching tag to participate.

### Models

Cross-task payloads are Pydantic models in [tidal/models/transcode.py](tidal/models/transcode.py): `TranscodeJobInput`, `CodecConfig`, `VideoResolution`, `ProbeResult`, `SegmentResult`, `VMAFResult`. Validators enforce even dimensions, CRF 0–51, supported containers (`mp4|mkv|webm|mov`), existing source paths.

### FFmpeg layer

[tidal/utilities/ffmpeg.py](tidal/utilities/ffmpeg.py) wraps invocations via `FFmpegProcessor` with a `ProgressData` callback consumed by `safe_create_progress` / `safe_update_progress` ([tidal/utilities/logging.py](tidal/utilities/logging.py)). Route ffmpeg calls through `FFmpegProcessor` instead of raw `subprocess` so progress + command construction stay consistent.

Dockerfile pins a specific FFmpeg release (BtbN GPL, libvmaf bundled). Bumping the version means editing both the `n<MAJOR.MINOR>` tag and the `-gpl-<MAJOR.MINOR>` filename suffix; ARM64 and AMD64 use the same version string.

## Tests

`tests/unit/` mocks `FFmpegProcessor` via the `mock_ffmpeg_processor` fixture. `tests/integration/` runs real ffmpeg via session-scoped fixtures (`sample_video`, `sample_video_no_audio`, `sample_audio`, `sample_chunk`) defined in [tests/conftest.py](tests/conftest.py) — they `pytest.skip` if ffmpeg not on PATH. Mark new integration tests with `@pytest.mark.integration`; mark long-running ones `@pytest.mark.slow`.

## Style

Ruff enforces tabs (indent-width=1, `tab` indent style) and double-quoted strings — preserve when editing. `E501` ignored despite line-length=120. Mypy runs with `disallow_incomplete_defs=True` and `ignore_missing_imports=True`; pydantic plugin enabled.
