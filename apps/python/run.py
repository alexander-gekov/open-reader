#!/usr/bin/env python3
"""
Simple script to run the Open Reader Python Backend.
"""

import os
import sys
import uvicorn
from dotenv import load_dotenv

def main():
    # Load environment variables
    load_dotenv()
    
    # Configuration
    host = os.getenv('HOST', '0.0.0.0')
    port = int(os.getenv('PORT', '8000'))
    log_level = os.getenv('LOG_LEVEL', 'info').lower()
    
    # Development mode check
    is_dev = '--dev' in sys.argv or os.getenv('ENVIRONMENT') == 'development'
    
    print(f"Starting Open Reader Python Backend...")
    print(f"Host: {host}")
    print(f"Port: {port}")
    print(f"Log Level: {log_level}")
    print(f"Development Mode: {is_dev}")
    print(f"API Documentation: http://{host}:{port}/docs")
    
    # Run the server
    uvicorn.run(
        "main:app",
        host=host,
        port=port,
        log_level=log_level,
        reload=is_dev,
        reload_dirs=["./"] if is_dev else None
    )

if __name__ == "__main__":
    main()