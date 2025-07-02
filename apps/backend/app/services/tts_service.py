import requests
import io
import tempfile
import os
from typing import Optional
from together import Together
from app.core.config import settings


class TTSService:
    def __init__(self):
        self.together_client = Together(api_key=settings.together_api_key)
    
    async def generate_speech_together_sonic(self, text: str, voice: str = "helpful woman") -> Optional[bytes]:
        """Generate speech using Together AI's Audio API with Cartesia Sonic"""
        try:
            # Use Together AI's Audio API to access Cartesia Sonic
            response = self.together_client.audio.speech.create(
                model="cartesia/sonic",
                input=text,
                voice=voice
            )
            
            # Create a temporary file to store the audio
            with tempfile.NamedTemporaryFile(delete=False, suffix=".mp3") as temp_file:
                temp_path = temp_file.name
            
            # Stream the response to the temporary file
            response.stream_to_file(temp_path)
            
            # Read the file content as bytes
            with open(temp_path, 'rb') as f:
                audio_data = f.read()
            
            # Clean up the temporary file
            os.unlink(temp_path)
            
            return audio_data
                
        except Exception as e:
            print(f"Error generating speech with Together AI Sonic: {e}")
            return None
    
    async def generate_speech_cartesia_direct(self, text: str, voice_id: str = "a0e99841-438c-4a64-b679-ae501e7d6091") -> Optional[bytes]:
        """Generate speech using Cartesia API directly (fallback)"""
        if not settings.cartesia_api_key:
            print("Cartesia API key not configured")
            return None
        
        try:
            url = "https://api.cartesia.ai/tts/bytes"
            headers = {
                "Cartesia-Version": "2024-06-10",
                "X-API-Key": settings.cartesia_api_key,
                "Content-Type": "application/json"
            }
            
            data = {
                "model_id": "sonic-english",
                "transcript": text,
                "voice": {
                    "mode": "id",
                    "id": voice_id
                },
                "output_format": {
                    "container": "mp3",
                    "encoding": "mp3",
                    "sample_rate": 22050
                }
            }
            
            response = requests.post(url, json=data, headers=headers)
            
            if response.status_code == 200:
                return response.content
            else:
                print(f"Cartesia API error: {response.status_code} - {response.text}")
                return None
                
        except Exception as e:
            print(f"Error generating speech with Cartesia direct: {e}")
            return None
    
    async def generate_speech_eleven_labs_demo(self, text: str) -> Optional[bytes]:
        """Generate speech using a demo TTS service (simulated)"""
        try:
            url = "https://api.elevenlabs.io/v1/text-to-speech/21m00Tcm4TlvDq8ikWAM"
            headers = {
                "Accept": "audio/mpeg",
                "Content-Type": "application/json",
                "xi-api-key": "demo_key"  # This would be a real API key
            }
            
            data = {
                "text": text,
                "model_id": "eleven_monolingual_v1",
                "voice_settings": {
                    "stability": 0.5,
                    "similarity_boost": 0.5
                }
            }
            
            response = requests.post(url, json=data, headers=headers)
            
            if response.status_code == 200:
                return response.content
            else:
                return None
                
        except Exception as e:
            print(f"Error with demo TTS: {e}")
            return self._generate_dummy_audio()
    
    def _generate_dummy_audio(self) -> bytes:
        """Generate a dummy audio file for testing"""
        return b"dummy_audio_content_for_testing"
    
    async def generate_speech(self, text: str, voice: str = "helpful woman") -> Optional[bytes]:
        """Generate speech using the best available service"""
        # Primary: Together AI Audio API with Cartesia Sonic
        audio_data = await self.generate_speech_together_sonic(text, voice)
        
        # Fallback 1: Direct Cartesia API (using voice ID instead of voice name)
        if not audio_data:
            audio_data = await self.generate_speech_cartesia_direct(text)
        
        # Fallback 2: Demo/dummy audio
        if not audio_data:
            audio_data = await self.generate_speech_eleven_labs_demo(text)
        
        return audio_data


tts_service = TTSService()