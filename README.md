Tidal is an application that helps simplify video encoding. Tidal allows users to queue, monitor, manage, and automate video transcoding.

Features
- Tidal is built using Golang as the primary language. The backend is a CLI and Rest API using the Echo framework. The API routes are located at /api
- The Tidal UI is built using Vue as a Single Page App. This is the "ui" folder. It is served at the root of the REST api server. "/"
- The application is deployed with at least 1 central server which is connected to Redis.
- Job processing and tracking is managed by the go asyncq package.
- Jobs run by default using Kubernetes batch jobs. This runs jobs in a separate container providing good isolation and scalability. Asynq is used to communicate job progress, status, and metadata.
- Any other other communiucation between workers and server occurs over the REST api
- A helm chart is provided in this repo for easy deployment
- The main dockerfile contains ffmpeg, tidal, and other utilities. It runs in server mode by default
- Postgresql database is used for application state like rules and automations. Asyncq uses redis for fast coordination, however, job status is written to postgresql for long term storage.

User interface
- The user interface is not authenticated
- Users can enqueue basic ffmpeg jobs.
- Users can create, manage, modify transcoding presets (with a few defaults provided). Users can delete the defaults if they so choose (seeded database by default)
- Users can create automations (when file is copied to folder X, run preset Y on file Z, then move source file Z to folder A and output transcoded file to folder B)
- Users can view job processing in realtime.
- Users can view job logs
- Users see a beautiful inteface written with tailwindcss
- Users can do all the things that asyncq enumerates


Couple of changes to this repo

- Use bun for ui
- Whole application should run via docker compose. When running locally, use bun vite server for ui, make it talk to the backend over http, should work from browser. Make application use worker container (not kube batch) when running with docker compose
- Remove .golangci.yml
- Remove sqlc.yaml
- Update .gitignore
- Organize and code into clean folder structure
- Write tests for backend, frontend (component tests). Wire up tests to run in CI
- Update CLAUDE.md
- Update Claude skills
- Create Changelog