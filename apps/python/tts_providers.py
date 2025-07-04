import os
import logging
import requests
import json
from abc import ABC, abstractmethod
from typing import Dict, Any, Optional
import boto3
from botocore.exceptions import ClientError
from elevenlabs import generate, Voice, VoiceSettings
import openai

logger = logging.getLogger(__name__)

class TTSProvider(ABC):
    """Abstract base class for TTS providers."""
    
    @abstractmethod
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate audio from text."""
        pass

class ElevenLabsProvider(TTSProvider):
    """ElevenLabs TTS provider."""
    
    def __init__(self, api_key: str):
        self.api_key = api_key
        os.environ["ELEVENLABS_API_KEY"] = api_key
    
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate audio using ElevenLabs API."""
        try:
            voice_id = options.get('voice', 'cgSgspJ2msm6clMCkdW9')
            model = options.get('model', 'eleven_flash_v2_5')
            
            # Use the elevenlabs library
            audio = generate(
                text=text,
                voice=Voice(
                    voice_id=voice_id,
                    settings=VoiceSettings(
                        stability=0.5,
                        similarity_boost=0.5
                    )
                ),
                model=model
            )
            
            return audio
            
        except Exception as e:
            logger.error(f"ElevenLabs TTS error: {str(e)}")
            raise ValueError(f"ElevenLabs TTS generation failed: {str(e)}")

class AWSPollyProvider(TTSProvider):
    """AWS Polly TTS provider."""
    
    def __init__(self, api_key: Optional[str] = None):
        self.polly_client = boto3.client(
            'polly',
            aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
            aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY'),
            region_name=os.getenv('AWS_REGION', 'us-east-1')
        )
    
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate audio using AWS Polly."""
        try:
            voice_id = options.get('voice', 'Joanna')
            engine = options.get('engine', 'neural')
            
            response = self.polly_client.synthesize_speech(
                Text=text,
                OutputFormat='mp3',
                VoiceId=voice_id,
                Engine=engine
            )
            
            audio_data = response['AudioStream'].read()
            return audio_data
            
        except ClientError as e:
            logger.error(f"AWS Polly TTS error: {str(e)}")
            raise ValueError(f"AWS Polly TTS generation failed: {str(e)}")

class OpenAIProvider(TTSProvider):
    """OpenAI TTS provider."""
    
    def __init__(self, api_key: str):
        self.client = openai.OpenAI(api_key=api_key)
    
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate audio using OpenAI TTS."""
        try:
            model = options.get('model', 'tts-1')
            voice = options.get('voice', 'alloy')
            
            response = self.client.audio.speech.create(
                model=model,
                voice=voice,
                input=text,
                response_format='mp3'
            )
            
            return response.content
            
        except Exception as e:
            logger.error(f"OpenAI TTS error: {str(e)}")
            raise ValueError(f"OpenAI TTS generation failed: {str(e)}")

class CartesiaProvider(TTSProvider):
    """Cartesia TTS provider (via Together.ai)."""
    
    def __init__(self, api_key: str):
        self.api_key = api_key
        self.base_url = "https://api.cartesia.ai/tts/bytes"
    
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate audio using Cartesia API."""
        try:
            model_id = options.get('model', 'sonic-english')
            voice_id = options.get('voice', 'a0e99841-438c-4a64-b679-ae501e7d6091')
            
            payload = {
                "model_id": model_id,
                "transcript": text,
                "voice": {
                    "mode": "id",
                    "id": voice_id
                },
                "output_format": {
                    "container": "mp3",
                    "bit_rate": 128000,
                    "sample_rate": 44100
                },
                "language": "en"
            }
            
            headers = {
                "Content-Type": "application/json",
                "Authorization": f"Bearer {self.api_key}",
                "Cartesia-Version": "2025-04-16"
            }
            
            response = requests.post(
                self.base_url,
                json=payload,
                headers=headers,
                timeout=30
            )
            
            if response.status_code != 200:
                raise ValueError(f"Cartesia API error: {response.status_code} - {response.text}")
            
            return response.content
            
        except Exception as e:
            logger.error(f"Cartesia TTS error: {str(e)}")
            raise ValueError(f"Cartesia TTS generation failed: {str(e)}")

class FallbackProvider(TTSProvider):
    """Fallback TTS provider using system TTS or simple generation."""
    
    def __init__(self, api_key: Optional[str] = None):
        pass
    
    async def generate_audio(self, text: str, options: Dict[str, Any]) -> bytes:
        """Generate fallback audio (placeholder)."""
        # This would implement a fallback TTS like gTTS or system TTS
        # For now, we'll raise an error to indicate fallback is needed
        raise ValueError("Fallback TTS provider not implemented yet")

class TTSProviderFactory:
    """Factory for creating TTS providers."""
    
    @staticmethod
    def create_provider(provider_name: str, api_key: str) -> TTSProvider:
        """Create a TTS provider instance."""
        
        if provider_name == "elevenlabs":
            return ElevenLabsProvider(api_key)
        elif provider_name == "polly":
            return AWSPollyProvider(api_key)
        elif provider_name == "openai":
            return OpenAIProvider(api_key)
        elif provider_name == "cartesia":
            return CartesiaProvider(api_key)
        elif provider_name == "fallback":
            return FallbackProvider(api_key)
        else:
            raise ValueError(f"Unsupported TTS provider: {provider_name}")
    
    @staticmethod
    def get_supported_providers() -> list:
        """Get list of supported TTS providers."""
        return ["elevenlabs", "polly", "openai", "cartesia", "fallback"]