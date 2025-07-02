#!/usr/bin/env python3
"""
Test script for Together AI + Cartesia Sonic TTS functionality
"""
import asyncio
import os
import sys
from pathlib import Path

# Add the app directory to Python path
sys.path.insert(0, str(Path(__file__).parent))

async def test_tts():
    """Test the TTS service with Together AI"""
    print("🎵 Testing Together AI + Cartesia Sonic TTS...")
    
    # Check if Together API key is set
    together_key = os.getenv('TOGETHER_API_KEY')
    if not together_key or together_key.startswith('your_'):
        print("❌ TOGETHER_API_KEY not set or using placeholder value")
        print("   Please set a real Together AI API key in your .env file")
        return
    
    try:
        # Import the TTS service
        from app.services.tts_service import TTSService
        
        # Create TTS service instance
        tts = TTSService()
        print("✅ TTS service created successfully")
        
        # Test text
        test_text = "Hello! This is a test of Together AI with Cartesia Sonic text-to-speech."
        print(f"📝 Testing with text: '{test_text}'")
        
        # Generate speech
        print("🔄 Generating speech...")
        audio_data = await tts.generate_speech_together_sonic(test_text)
        
        if audio_data:
            print(f"✅ Speech generated successfully! Audio size: {len(audio_data)} bytes")
            
            # Save to file for testing
            output_file = "test_speech.mp3"
            with open(output_file, 'wb') as f:
                f.write(audio_data)
            print(f"💾 Audio saved to {output_file}")
            
        else:
            print("❌ Failed to generate speech")
            print("   This could be due to:")
            print("   - Invalid API key")
            print("   - Network connectivity issues")
            print("   - API quota exceeded")
            print("   - Service temporarily unavailable")
            
    except ImportError as e:
        print(f"❌ Import error: {e}")
        print("   Make sure all dependencies are installed")
    except Exception as e:
        print(f"❌ Error: {e}")

def main():
    print("🧪 Together AI + Cartesia Sonic TTS Test")
    print("=" * 50)
    
    # Run the async test
    asyncio.run(test_tts())

if __name__ == "__main__":
    main()