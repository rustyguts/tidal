import subprocess
import asyncio
import os
import re
import json
import time
import logging
import tempfile
import shutil
import signal
from abc import ABC, abstractmethod
from contextlib import contextmanager, asynccontextmanager
from dataclasses import dataclass
from enum import Enum
from functools import wraps
from pathlib import Path
from typing import Dict, Any, List, Optional, Union, Callable, AsyncGenerator
from concurrent.futures import ThreadPoolExecutor

# Exception Hierarchy
class FFmpegError(Exception):
		"""Base exception for FFmpeg operations"""
		def __init__(self, message: str, exit_code: int = None, stderr: str = None):
				self.message = message
				self.exit_code = exit_code
				self.stderr = stderr
				super().__init__(message)

class FFmpegTimeoutError(FFmpegError):
		"""FFmpeg operation timeout"""
		pass

class FFmpegSecurityError(FFmpegError):
		"""Security-related FFmpeg error"""
		pass

class FFmpegProcessError(FFmpegError):
		"""Process execution error"""
		pass

# Progress Data Structure
@dataclass
class ProgressData:
		frame: Optional[int] = None
		fps: Optional[float] = None
		time_seconds: Optional[float] = None
		bitrate: Optional[int] = None
		speed: Optional[float] = None
		size: Optional[int] = None
		progress_percent: Optional[float] = None
		status: str = "processing"

# Event Types
class EventType(Enum):
		STARTED = "started"
		PROGRESS = "progress"
		COMPLETED = "completed"
		ERROR = "error"
		TERMINATED = "terminated"

class FFmpegProcessor:
		"""
		Comprehensive FFmpeg processor with modern Python best practices.
		
		Features:
		- Secure subprocess management
		- Real-time progress reporting
		- Async/sync support
		- Robust error handling
		- Resource management
		- Multiple output format support
		"""
		
		def __init__(self, 
								 ffmpeg_path: str = None,
								 timeout: int = 3600,
								 max_memory_mb: int = 2048,
								 temp_dir: str = None,
								 log_level: int = logging.INFO):
				
				# Configuration
				self.ffmpeg_path = self._get_safe_ffmpeg_path(ffmpeg_path)
				self.timeout = timeout
				self.max_memory_mb = max_memory_mb
				self.temp_dir = temp_dir or tempfile.gettempdir()
				
				# Setup logging
				self.logger = self._setup_logger(log_level)
				
				# Event handlers
				self.event_handlers: Dict[EventType, List[Callable]] = {
						event_type: [] for event_type in EventType
				}
				
				# Progress parsing patterns
				self.progress_patterns = {
						'frame': re.compile(r'frame=\s*(\d+)'),
						'fps': re.compile(r'fps=\s*([0-9.]+)'),
						'time': re.compile(r'out_time_ms=(\d+)'),
						'bitrate': re.compile(r'bitrate=\s*([0-9.]+)kbits/s'),
						'speed': re.compile(r'speed=\s*([0-9.]+)x'),
						'size': re.compile(r'total_size=(\d+)'),
						'progress': re.compile(r'progress=(\w+)')
				}
		
		def _get_safe_ffmpeg_path(self, custom_path: str = None) -> str:
				"""Get validated FFmpeg executable path"""
				if custom_path:
						if not os.path.isfile(custom_path) or not os.access(custom_path, os.X_OK):
								raise FFmpegSecurityError(f"Invalid FFmpeg path: {custom_path}")
						return custom_path
				
				# Search in trusted paths
				trusted_paths = ['/usr/bin/ffmpeg', '/usr/local/bin/ffmpeg', '/opt/homebrew/bin/ffmpeg']
				for path in trusted_paths:
						if os.path.isfile(path) and os.access(path, os.X_OK):
								return path
				
				# Fallback to system PATH
				ffmpeg_path = shutil.which('ffmpeg')
				if not ffmpeg_path:
						raise FFmpegError("FFmpeg executable not found")
				
				return ffmpeg_path
		
		def _setup_logger(self, log_level: int) -> logging.Logger:
				"""Setup structured logging"""
				logger = logging.getLogger(f"FFmpegProcessor_{id(self)}")
				logger.setLevel(log_level)
				
				if not logger.handlers:
						handler = logging.StreamHandler()
						formatter = logging.Formatter(
								'%(asctime)s - %(name)s - %(levelname)s - %(message)s'
						)
						handler.setFormatter(formatter)
						logger.addHandler(handler)
				
				return logger
		
		def _validate_arguments(self, args: List[str]) -> List[str]:
				"""Validate and sanitize FFmpeg arguments"""
				if not isinstance(args, list):
						raise FFmpegSecurityError("Arguments must be provided as a list")
				
				validated_args = []
				
				for arg in args:
						if not isinstance(arg, str):
								raise FFmpegSecurityError(f"All arguments must be strings, got: {type(arg)}")
						
						# Check for shell injection attempts
						dangerous_chars = [';', '&', '|', '`', '$', '(', ')', '<', '>']
						if any(char in arg for char in dangerous_chars):
								raise FFmpegSecurityError(f"Potentially dangerous characters in argument: {arg}")
						
						validated_args.append(arg)
				
				return validated_args
		
		def _parse_progress_line(self, line: str, total_duration: float = None) -> Optional[ProgressData]:
				"""Parse FFmpeg progress output line"""
				progress_data = ProgressData()
				
				for key, pattern in self.progress_patterns.items():
						match = pattern.search(line)
						if match:
								value = match.group(1)
								
								if key == 'frame':
										progress_data.frame = int(value)
								elif key == 'fps':
										progress_data.fps = float(value)
								elif key == 'time':
										# Convert microseconds to seconds
										progress_data.time_seconds = int(value) / 1_000_000
								elif key == 'bitrate':
										progress_data.bitrate = int(float(value) * 1000)  # Convert to bits/s
								elif key == 'speed':
										progress_data.speed = float(value)
								elif key == 'size':
										progress_data.size = int(value)
								elif key == 'progress':
										progress_data.status = value
				
				# Calculate percentage if we have duration
				if total_duration and progress_data.time_seconds:
						progress_data.progress_percent = min(
								(progress_data.time_seconds / total_duration) * 100, 100.0
						)
				
				return progress_data if any([
						progress_data.frame, progress_data.fps, progress_data.time_seconds
				]) else None
		
		def _get_media_duration(self, input_file: str) -> Optional[float]:
				"""Get media duration using ffprobe"""
				try:
						probe_cmd = [
								'ffprobe', '-v', 'quiet', '-print_format', 'json',
								'-show_format', input_file
						]
						
						result = subprocess.run(
								probe_cmd,
								capture_output=True,
								text=True,
								timeout=30
						)
						
						if result.returncode == 0:
								data = json.loads(result.stdout)
								return float(data['format']['duration'])
								
				except (subprocess.TimeoutExpired, json.JSONDecodeError, KeyError):
						pass
				
				return None
		
		@contextmanager
		def _managed_temp_file(self, suffix: str = '.tmp'):
				"""Context manager for temporary file management"""
				temp_file = tempfile.NamedTemporaryFile(
						delete=False, 
						suffix=suffix,
						dir=self.temp_dir
				)
				temp_file.close()
				
				try:
						yield temp_file.name
				finally:
						try:
								os.unlink(temp_file.name)
						except OSError:
								pass
		
		def on(self, event_type: EventType, handler: Callable):
				"""Register event handler"""
				self.event_handlers[event_type].append(handler)
		
		def _emit_event(self, event_type: EventType, data: Any = None):
				"""Emit event to registered handlers"""
				for handler in self.event_handlers[event_type]:
						try:
								handler(data)
						except Exception as e:
								self.logger.warning(f"Event handler error: {e}")
		
		def execute_sync(self, 
										args: List[str],
										input_file: str = None,
										capture_output: bool = True,
										progress_callback: Callable[[ProgressData], None] = None,
										**kwargs) -> Dict[str, Any]:
				"""
				Synchronous FFmpeg execution with progress monitoring.
				
				Args:
						args: FFmpeg arguments (without 'ffmpeg' prefix)
						input_file: Input file path for duration calculation
						capture_output: Whether to capture stdout/stderr
						progress_callback: Callback function for progress updates
						**kwargs: Additional subprocess arguments
						
				Returns:
						Dict with execution results including stdout, stderr, return_code
				"""
				
				# Validate arguments
				validated_args = self._validate_arguments(args)
				command = [self.ffmpeg_path] + validated_args
				
				self.logger.info(f"Executing FFmpeg: {' '.join(command)}")
				
				# Get duration for progress calculation
				duration = None
				if input_file and progress_callback:
						duration = self._get_media_duration(input_file)
				
				# Setup progress monitoring
				with self._managed_temp_file('.progress') as progress_file:
						# Add progress reporting if callback provided
						if progress_callback:
								command.extend(['-progress', progress_file])
						
						# Emit start event
						self._emit_event(EventType.STARTED, {'command': command})
						
						try:
								# Start process
								process = subprocess.Popen(
										command,
										stdout=subprocess.PIPE if capture_output else None,
										stderr=subprocess.PIPE if capture_output else None,
										universal_newlines=True,
										**kwargs
								)
								
								# Monitor progress if callback provided
								if progress_callback:
										self._monitor_progress_sync(
												process, progress_file, duration, progress_callback
										)
								
								# Wait for completion
								stdout, stderr = process.communicate(timeout=self.timeout)
								
								result = {
										'returncode': process.returncode,
										'stdout': stdout,
										'stderr': stderr,
										'command': command
								}
								
								if process.returncode == 0:
										self._emit_event(EventType.COMPLETED, result)
								else:
										error = FFmpegProcessError(
												f"FFmpeg failed with exit code {process.returncode}",
												process.returncode,
												stderr
										)
										self._emit_event(EventType.ERROR, error)
										raise error
								
								return result
								
						except subprocess.TimeoutExpired:
								process.kill()
								process.wait()
								error = FFmpegTimeoutError(f"FFmpeg timeout after {self.timeout}s")
								self._emit_event(EventType.ERROR, error)
								raise error
						
						except Exception as e:
								if process.poll() is None:
										process.terminate()
										process.wait()
								self._emit_event(EventType.ERROR, e)
								raise
		
		def _monitor_progress_sync(self, 
															process: subprocess.Popen, 
															progress_file: str,
															duration: float,
															callback: Callable[[ProgressData], None]):
				"""Monitor progress synchronously"""
				last_size = 0
				
				while process.poll() is None:
						try:
								# Read progress file
								if os.path.exists(progress_file):
										current_size = os.path.getsize(progress_file)
										if current_size > last_size:
												with open(progress_file, 'r') as f:
														content = f.read()
														
												# Parse progress data
												for line in content.strip().split('\n'):
														if '=' in line:
																progress_data = self._parse_progress_line(line, duration)
																if progress_data:
																		callback(progress_data)
																		self._emit_event(EventType.PROGRESS, progress_data)
												
												last_size = current_size
								
								time.sleep(0.1)  # Small delay to prevent excessive file reading
								
						except (OSError, IOError):
								continue
		
		async def execute_async(self,
													 args: List[str],
													 input_file: str = None,
													 capture_output: bool = True,
													 progress_callback: Optional[Callable[[ProgressData], None]] = None,
													 **kwargs) -> Dict[str, Any]:
				"""
				Asynchronous FFmpeg execution with progress monitoring.
				
				Args:
						args: FFmpeg arguments (without 'ffmpeg' prefix)
						input_file: Input file path for duration calculation
						capture_output: Whether to capture stdout/stderr
						progress_callback: Async callback function for progress updates
						**kwargs: Additional asyncio subprocess arguments
						
				Returns:
						Dict with execution results including stdout, stderr, return_code
				"""
				
				# Validate arguments
				validated_args = self._validate_arguments(args)
				command = [self.ffmpeg_path] + validated_args
				
				self.logger.info(f"Executing FFmpeg async: {' '.join(command)}")
				
				# Get duration for progress calculation
				duration = None
				if input_file and progress_callback:
						duration = await self._get_media_duration_async(input_file)
				
				# Setup progress monitoring
				with self._managed_temp_file('.progress') as progress_file:
						# Add progress reporting if callback provided
						if progress_callback:
								command.extend(['-progress', progress_file])
						
						# Emit start event
						self._emit_event(EventType.STARTED, {'command': command})
						
						try:
								# Start async process
								process = await asyncio.create_subprocess_exec(
										*command,
										stdout=asyncio.subprocess.PIPE if capture_output else None,
										stderr=asyncio.subprocess.PIPE if capture_output else None,
										**kwargs
								)
								
								# Monitor progress if callback provided
								progress_task = None
								if progress_callback:
										progress_task = asyncio.create_task(
												self._monitor_progress_async(
														process, progress_file, duration, progress_callback
												)
										)
								
								# Wait for completion with timeout
								try:
										stdout, stderr = await asyncio.wait_for(
												process.communicate(), 
												timeout=self.timeout
										)
										
										if progress_task:
												progress_task.cancel()
										
										# Decode bytes to string if needed
										if stdout:
												stdout = stdout.decode('utf-8')
										if stderr:
												stderr = stderr.decode('utf-8')
										
										result = {
												'returncode': process.returncode,
												'stdout': stdout,
												'stderr': stderr,
												'command': command
										}
										
										if process.returncode == 0:
												self._emit_event(EventType.COMPLETED, result)
										else:
												error = FFmpegProcessError(
														f"FFmpeg failed with exit code {process.returncode}",
														process.returncode,
														stderr
												)
												self._emit_event(EventType.ERROR, error)
												raise error
										
										return result
										
								except asyncio.TimeoutError:
										process.kill()
										await process.wait()
										if progress_task:
												progress_task.cancel()
										error = FFmpegTimeoutError(f"FFmpeg timeout after {self.timeout}s")
										self._emit_event(EventType.ERROR, error)
										raise error
								
						except Exception as e:
								if process.returncode is None:
										process.terminate()
										await process.wait()
								self._emit_event(EventType.ERROR, e)
								raise
		
		async def _get_media_duration_async(self, input_file: str) -> Optional[float]:
				"""Get media duration asynchronously using ffprobe"""
				try:
						process = await asyncio.create_subprocess_exec(
								'ffprobe', '-v', 'quiet', '-print_format', 'json',
								'-show_format', input_file,
								stdout=asyncio.subprocess.PIPE,
								stderr=asyncio.subprocess.PIPE
						)
						
						stdout, _ = await asyncio.wait_for(process.communicate(), timeout=30)
						
						if process.returncode == 0:
								data = json.loads(stdout.decode('utf-8'))
								return float(data['format']['duration'])
								
				except (asyncio.TimeoutError, json.JSONDecodeError, KeyError):
						pass
				
				return None
		
		async def _monitor_progress_async(self,
																		process: asyncio.subprocess.Process,
																		progress_file: str,
																		duration: float,
																		callback: Callable[[ProgressData], None]):
				"""Monitor progress asynchronously"""
				last_size = 0
				
				while process.returncode is None:
						try:
								# Read progress file
								if os.path.exists(progress_file):
										current_size = os.path.getsize(progress_file)
										if current_size > last_size:
												with open(progress_file, 'r') as f:
														content = f.read()
												
												# Parse progress data
												for line in content.strip().split('\n'):
														if '=' in line:
																progress_data = self._parse_progress_line(line, duration)
																if progress_data:
																		if asyncio.iscoroutinefunction(callback):
																				await callback(progress_data)
																		else:
																				callback(progress_data)
																		self._emit_event(EventType.PROGRESS, progress_data)
												
												last_size = current_size
								
								await asyncio.sleep(0.1)  # Non-blocking delay
								
						except (OSError, IOError):
								continue
						except asyncio.CancelledError:
								break
		
		def convert_video(self,
										 input_file: str,
										 output_file: str,
										 **options) -> Dict[str, Any]:
				"""
				High-level video conversion method.
				
				Args:
						input_file: Input video file path
						output_file: Output video file path
						**options: FFmpeg options (e.g., codec='libx264', preset='medium')
						
				Returns:
						Dict with conversion results
				"""
				
				# Build FFmpeg arguments
				args = ['-i', input_file]
				
				# Add options
				for key, value in options.items():
						if key.startswith('codec'):
								args.extend([f'-c:{key.split(":")[-1]}', value])
						else:
								args.extend([f'-{key}', str(value)])
				
				# Add output file
				args.append(output_file)
				
				# Add overwrite flag
				args.insert(0, '-y')
				
				return self.execute_sync(args, input_file)
		
		async def convert_video_async(self,
																 input_file: str,
																 output_file: str,
																 **options) -> Dict[str, Any]:
				"""Async version of convert_video"""
				
				# Build FFmpeg arguments (same as sync version)
				args = ['-i', input_file]
				
				for key, value in options.items():
						if key.startswith('codec'):
								args.extend([f'-c:{key.split(":")[-1]}', value])
						else:
								args.extend([f'-{key}', str(value)])
				
				args.append(output_file)
				args.insert(0, '-y')
				
				return await self.execute_async(args, input_file)

# Convenience Functions for Common Use Cases

def simple_convert(input_file: str, 
									output_file: str,
									codec: str = 'libx264',
									preset: str = 'medium',
									progress_callback: Callable[[ProgressData], None] = None) -> Dict[str, Any]:
		"""
		Simple video conversion function.
		
		Args:
				input_file: Input video file
				output_file: Output video file
				codec: Video codec (default: libx264)
				preset: Encoding preset (default: medium)
				progress_callback: Optional progress callback
				
		Returns:
				Conversion results
		"""
		processor = FFmpegProcessor()
		
		if progress_callback:
				processor.on(EventType.PROGRESS, progress_callback)
		
		return processor.convert_video(
				input_file, 
				output_file,
				codec=codec,
				preset=preset
		)

async def simple_convert_async(input_file: str,
															output_file: str,
															codec: str = 'libx264',
															preset: str = 'medium',
															progress_callback: Callable[[ProgressData], None] = None) -> Dict[str, Any]:
		"""Async version of simple_convert"""
		processor = FFmpegProcessor()
		
		if progress_callback:
				processor.on(EventType.PROGRESS, progress_callback)
		
		return await processor.convert_video_async(
				input_file,
				output_file,
				codec=codec,
				preset=preset
		)

# Batch Processing Utilities

class BatchProcessor:
		"""Utility for processing multiple files concurrently"""
		
		def __init__(self, max_concurrent: int = 4):
				self.max_concurrent = max_concurrent
				self.processor = FFmpegProcessor()
		
		async def process_batch_async(self,
																 files: List[tuple],
																 conversion_options: Dict[str, Any] = None) -> List[Dict[str, Any]]:
				"""
				Process multiple files asynchronously.
				
				Args:
						files: List of (input_file, output_file) tuples
						conversion_options: Common conversion options
						
				Returns:
						List of conversion results
				"""
				semaphore = asyncio.Semaphore(self.max_concurrent)
				options = conversion_options or {}
				
				async def process_single(input_file: str, output_file: str):
						async with semaphore:
								return await self.processor.convert_video_async(
										input_file, output_file, **options
								)
				
				tasks = [
						process_single(input_file, output_file)
						for input_file, output_file in files
				]
				
				return await asyncio.gather(*tasks, return_exceptions=True)