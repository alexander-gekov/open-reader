import boto3
from botocore.config import Config
from typing import Optional
import io
from app.core.config import settings


class S3Service:
    def __init__(self):
        self.client = boto3.client(
            "s3",
            aws_access_key_id=settings.r2_access_key_id,
            aws_secret_access_key=settings.r2_secret_access_key,
            endpoint_url=settings.r2_endpoint,
            config=Config(signature_version="v4"),
            region_name="auto",
        )
        self.bucket_name = settings.r2_bucket_name
    
    def upload_file(self, file_content: bytes, key: str, content_type: str = "application/octet-stream") -> bool:
        """Upload file to R2 storage"""
        try:
            self.client.put_object(
                Bucket=self.bucket_name,
                Key=key,
                Body=file_content,
                ContentType=content_type
            )
            return True
        except Exception as e:
            print(f"Error uploading file: {e}")
            return False
    
    def download_file(self, key: str) -> Optional[bytes]:
        """Download file from R2 storage"""
        try:
            response = self.client.get_object(Bucket=self.bucket_name, Key=key)
            return response["Body"].read()
        except Exception as e:
            print(f"Error downloading file: {e}")
            return None
    
    def get_presigned_url(self, key: str, expiration: int = 3600) -> Optional[str]:
        """Generate presigned URL for file access"""
        try:
            url = self.client.generate_presigned_url(
                "get_object",
                Params={"Bucket": self.bucket_name, "Key": key},
                ExpiresIn=expiration
            )
            return url
        except Exception as e:
            print(f"Error generating presigned URL: {e}")
            return None
    
    def delete_file(self, key: str) -> bool:
        """Delete file from R2 storage"""
        try:
            self.client.delete_object(Bucket=self.bucket_name, Key=key)
            return True
        except Exception as e:
            print(f"Error deleting file: {e}")
            return False
    
    def file_exists(self, key: str) -> bool:
        """Check if file exists in R2 storage"""
        try:
            self.client.head_object(Bucket=self.bucket_name, Key=key)
            return True
        except:
            return False


s3_service = S3Service()