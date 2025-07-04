#!/usr/bin/env python3
"""
Development scripts for the Open Reader Python Backend.
Usage: python scripts.py <command>

Commands:
  install     - Install dependencies
  migrate     - Run database migrations
  dev         - Start development server
  test        - Run tests
  format      - Format code
  lint        - Run linting
  build       - Build Docker image
  clean       - Clean up temporary files
"""

import os
import sys
import subprocess
from pathlib import Path

def run_command(cmd, description=""):
    """Run a shell command."""
    if description:
        print(f"▶️  {description}")
    print(f"💻 {cmd}")
    result = subprocess.run(cmd, shell=True)
    if result.returncode != 0:
        print(f"❌ Command failed: {cmd}")
        sys.exit(1)
    print("✅ Done\n")

def install():
    """Install dependencies."""
    run_command("pip install -r requirements.txt", "Installing Python dependencies")
    run_command("python -c 'import nltk; nltk.download(\"punkt\", quiet=True)'", "Downloading NLTK data")

def migrate():
    """Run database migrations."""
    run_command("alembic upgrade head", "Running database migrations")

def dev():
    """Start development server."""
    run_command("python run.py --dev", "Starting development server")

def test():
    """Run tests."""
    if not Path("tests").exists():
        print("ℹ️  No tests directory found. Creating basic test structure...")
        Path("tests").mkdir(exist_ok=True)
        Path("tests/__init__.py").touch()
        with open("tests/test_main.py", "w") as f:
            f.write("""
import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

def test_health_check():
    response = client.get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "healthy"}
""")
    run_command("pytest", "Running tests")

def format_code():
    """Format code."""
    run_command("black .", "Formatting code with Black")
    run_command("isort .", "Sorting imports with isort")

def lint():
    """Run linting."""
    run_command("flake8 .", "Running flake8 linting")
    run_command("mypy . --ignore-missing-imports", "Running mypy type checking")

def build():
    """Build Docker image."""
    run_command("docker build -t open-reader-python .", "Building Docker image")

def clean():
    """Clean up temporary files."""
    run_command("find . -type d -name '__pycache__' -exec rm -rf {} + 2>/dev/null || true", "Cleaning __pycache__ directories")
    run_command("find . -name '*.pyc' -delete 2>/dev/null || true", "Cleaning .pyc files")
    run_command("find . -name '.pytest_cache' -exec rm -rf {} + 2>/dev/null || true", "Cleaning pytest cache")

def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)
    
    command = sys.argv[1]
    
    commands = {
        'install': install,
        'migrate': migrate,
        'dev': dev,
        'test': test,
        'format': format_code,
        'lint': lint,
        'build': build,
        'clean': clean,
    }
    
    if command not in commands:
        print(f"❌ Unknown command: {command}")
        print(__doc__)
        sys.exit(1)
    
    commands[command]()

if __name__ == "__main__":
    main()