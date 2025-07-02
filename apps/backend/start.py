#!/usr/bin/env python3
import os
import sys
from pathlib import Path

def check_environment():
    """Check if required environment variables are set"""
    required_vars = [
        'DATABASE_URL',
        'REDIS_URL', 
        'R2_ACCOUNT_ID',
        'R2_ACCESS_KEY_ID',
        'R2_SECRET_ACCESS_KEY',
        'R2_ENDPOINT',
        'R2_BUCKET_NAME',
        'TOGETHER_API_KEY'
    ]
    
    missing_vars = []
    for var in required_vars:
        if not os.getenv(var) or os.getenv(var).startswith('your_'):
            missing_vars.append(var)
    
    if missing_vars:
        print("❌ Missing or invalid environment variables:")
        for var in missing_vars:
            print(f"   - {var}")
        print("\n📝 Please update your .env file with real values")
        print("   Copy .env.example to .env and fill in your actual credentials")
        return False
    
    return True

def main():
    print("🚀 Starting Open Reader Backend...")
    
    # Check if .env file exists
    if not Path('.env').exists():
        print("📝 Creating .env file from .env.example...")
        import shutil
        shutil.copy('.env.example', '.env')
        print("✅ .env file created. Please edit it with your actual credentials.")
        return
    
    # Check environment variables
    if not check_environment():
        return
    
    print("✅ Environment variables look good!")
    print("🎵 Starting FastAPI server with Together AI + Cartesia Sonic TTS...")
    
    # Start the server
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        reload_dirs=["app"]
    )

if __name__ == "__main__":
    main()