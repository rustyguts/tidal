# https://typer.tiangolo.com/

import asyncio
import os
import re
import time

import requests
import typer
from prefect import serve
from prefect.client.orchestration import get_client
from prefect.client.schemas.actions import GlobalConcurrencyLimitCreate, GlobalConcurrencyLimitUpdate
from prefect.client.schemas.objects import ConcurrencyLimitConfig
from prefect.variables import Variable
from rich import print
from rich.console import Console

from tidal.flows.transcode import transcode
from tidal.flows.chunked_transcode import chunked_transcode
from tidal.utils.types import Environment
from tidal.utils.vars import GlobalQueues, TaskQueues

console = Console()
app = typer.Typer(no_args_is_help=True)

EnvironmentFlag = typer.Option(help="The deployment environment")

PREFECT_APP_VARIABLE_KEY = "app_tidal"


def get_application_version() -> str:
	with open("pyproject.toml") as f:
		content = f.read()

	match = re.search(r'\[project\].*?version\s*=\s*"([^"]+)"', content, re.DOTALL)
	if match:
		return match.group(1)
	raise ValueError("Could not find version in pyproject.toml")


@app.command()
async def prefect_concurrency_limits():
	print("[bold purple]Setting Prefect Concurrency Limits[/bold purple]")

	if not prefect_server_check_configuration():
		print("[bold red]Cannot proceed without Prefect server![/bold red]")
		raise typer.Abort()

	async with get_client() as client:
		for queue in TaskQueues:
			print(f"[blue]Setting Prefect Concurrency Limit for {queue}[/blue]")
			try:
				await client.create_concurrency_limit(tag=queue.name, concurrency_limit=queue.limit)
				await client.read_concurrency_limit_by_tag(tag=queue.name)
				print(f"[green]Successfully set Prefect Concurrency Limit for {queue}![/green]")
			except Exception as e:
				print(f"[bold red]Failed to set concurrency limit for {queue}: {e}[/bold red]")

		for queue in GlobalQueues:
			print(f"[blue]Setting Prefect Concurrency Limit for {queue.name}[/blue]")
			try:
				await client.read_global_concurrency_limit_by_name(name=queue.name)
				await client.update_global_concurrency_limit(
					name=queue.name,
					concurrency_limit=GlobalConcurrencyLimitUpdate(
						name=queue.name,
						limit=queue.limit,
						slot_decay_per_second=queue.slot_decay_per_second,
					),
				)

				print(f"[green]Successfully set Prefect Concurrency Limit for {queue.name}![/green]")
			except Exception:
				print(f"[green]Queue did not exist, creating now... {queue.name}![/green]")
				await client.create_global_concurrency_limit(
					concurrency_limit=GlobalConcurrencyLimitCreate(
						limit=queue.limit, name=queue.name, slot_decay_per_second=queue.slot_decay_per_second
					)
				)


@app.command()
def prefect_variables(environment: Environment = EnvironmentFlag) -> None:
	print("[bold purple]Setting Prefect Variables[/bold purple]")

	if not prefect_server_check_configuration():
		print("[bold red]Cannot proceed without Prefect server![/bold red]")
		raise typer.Abort()

	prefect_variable = {
		"test_var": os.getenv("TEST_VAR", "test_var_default"),
	}

	Variable.set(PREFECT_APP_VARIABLE_KEY, prefect_variable, overwrite=True)
	print("[bold green]Successfully set Prefect variables![/bold green]")


def prefect_server_check_configuration(max_retries: int = 5, delay: int = 2) -> bool:
	api_url = os.getenv("PREFECT_API_URL", "http://localhost:4200/api")
	print(f"[bold blue]PREFECT_API_URL was sourced from env: {api_url}[/bold blue]")

	if not api_url.endswith("/api"):
		raise ValueError("Prefect API URL must end with '/api', I learned this the hard way...")

	for _ in range(max_retries):
		try:
			response = requests.get(f"{api_url.rstrip('/api')}/api/health")
			if response.status_code == 200:
				return True
			else:
				print(f"[yellow]Prefect server not ready yet (status: {response.status_code}), retrying...[/yellow]")
		except Exception as e:
			print(f"[yellow]Prefect server connection error: {e}, retrying...[/yellow]")

		time.sleep(delay)

	print("[bold red]Failed to connect to Prefect server after multiple attempts[/bold red]")
	return False


@app.command()
def prefect_flows(
	environment: Environment = EnvironmentFlag,
) -> None:
	app_version = get_application_version()
	deployment_tags = [environment.value, app_version]

	if not prefect_server_check_configuration():
		print("[bold red]Cannot proceed without Prefect server![/bold red]")
		raise typer.Abort()

	print("[bold purple]Serving Local Prefect Flows In Process[/bold purple]")
	chunked_transcode_deployment = chunked_transcode.to_deployment(
		name="chunked-transcode",
		version=app_version,
		tags=deployment_tags,
		concurrency_limit=ConcurrencyLimitConfig(limit=4),
	)

	transcode_deployment = transcode.to_deployment(
		name="transcode",
		version=app_version,
		tags=deployment_tags,
		concurrency_limit=ConcurrencyLimitConfig(limit=4),
	)

	serve(chunked_transcode_deployment, transcode_deployment)  # type: ignore
	print("[bold green]Successfully deployed Prefect Flows![/bold green]")


@app.command()
def all(
	environment: Environment = EnvironmentFlag,
) -> None:
	print(f"🌊 [bold purple]Deploying Tidal to [green]{environment}[/green] environment![/bold purple] 🌊")
	prefect_variables(environment)
	asyncio.run(prefect_concurrency_limits())
	prefect_flows(environment)
	print("🟢 [bold purple]Deployment Success![/bold purple] 🟢")


if __name__ == "__main__":
	app()
