from pathlib import Path

import pytest

from tidal.models.transcode import CodecConfig
from tidal.tasks.audio_transcode import transcode_audio, _codec_extension


class TestAudioTranscode:
	def test_transcode_audio_aac(self, sample_audio, temp_dir, ffmpeg_available):
		"""Test transcoding audio to AAC."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(audio_codec="aac", audio_bitrate="128k")

		result = transcode_audio.fn(
			audio_path=sample_audio,
			output_dir=str(temp_dir),
			codec=codec,
		)

		assert Path(result).exists()
		assert Path(result).suffix == ".m4a"
		assert Path(result).stat().st_size > 0

	def test_transcode_audio_creates_output_dir(self, sample_audio, temp_dir, ffmpeg_available):
		"""Test that audio transcode creates output directory."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(audio_codec="aac", audio_bitrate="64k")
		nested_dir = temp_dir / "nested" / "audio"

		result = transcode_audio.fn(
			audio_path=sample_audio,
			output_dir=str(nested_dir),
			codec=codec,
		)

		assert Path(result).exists()

	def test_transcode_audio_override_codec(self, sample_audio, temp_dir, ffmpeg_available):
		"""Test overriding the codec via audio_codec parameter."""
		if not ffmpeg_available:
			pytest.skip("FFmpeg not available")

		codec = CodecConfig(audio_codec="aac", audio_bitrate="64k")

		result = transcode_audio.fn(
			audio_path=sample_audio,
			output_dir=str(temp_dir),
			codec=codec,
			audio_codec="libmp3lame",
		)

		assert Path(result).exists()
		assert Path(result).suffix == ".mp3"


class TestCodecExtension:
	def test_aac_extension(self):
		assert _codec_extension("aac") == "m4a"

	def test_opus_extension(self):
		assert _codec_extension("libopus") == "opus"
		assert _codec_extension("opus") == "opus"

	def test_mp3_extension(self):
		assert _codec_extension("libmp3lame") == "mp3"
		assert _codec_extension("mp3") == "mp3"

	def test_flac_extension(self):
		assert _codec_extension("flac") == "flac"

	def test_vorbis_extension(self):
		assert _codec_extension("libvorbis") == "ogg"

	def test_wav_extension(self):
		assert _codec_extension("pcm_s16le") == "wav"

	def test_unknown_codec_fallback(self):
		assert _codec_extension("unknown_codec") == "mka"
