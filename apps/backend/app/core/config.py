from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    database_url: str
    redis_url: str = "redis://localhost:6379"
    
    r2_account_id: str
    r2_access_key_id: str
    r2_secret_access_key: str
    r2_endpoint: str
    r2_bucket_name: str
    
    together_api_key: str
    cartesia_api_key: Optional[str] = None
    
    app_name: str = "Open Reader Backend"
    debug: bool = False

    class Config:
        env_file = ".env"
        case_sensitive = False


settings = Settings()