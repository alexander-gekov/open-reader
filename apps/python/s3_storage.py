import boto3
import os
import logging
from typing import Optional
from botocore.exceptions import ClientError
from datetime import datetime

logger = logging.getLogger(__name__)

class S3Storage:
    def __init__(self):
        self.s3_client = boto3.client(
            's3',
            aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
            aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY'),
            region_name=os.getenv('AWS_REGION', 'us-east-1')
        )
        self.bucket_name = os.getenv('AWS_BUCKET_NAME', 'open-reader')
    
    async def upload_file(self, file_content: bytes, key: str, content_type: str = 'application/octet-stream') -> str:
        """
        Upload a file to S3 and return the URL.
        """
        try:
            # Add timestamp to make keys unique
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            unique_key = f"{timestamp}_{key}"
            
            self.s3_client.put_object(
                Bucket=self.bucket_name,
                Key=unique_key,
                Body=file_content,
                ContentType=content_type
            )
            
            # Return the S3 URL
            url = f"https://{self.bucket_name}.s3.amazonaws.com/{unique_key}"
            logger.info(f"Successfully uploaded file to S3: {url}")
            return url
            
        except ClientError as e:
            logger.error(f"Error uploading file to S3: {str(e)}")
            raise ValueError(f"Failed to upload file: {str(e)}")
    
    async def upload_audio(self, audio_data: bytes, filename: str, chunk_index: int) -> str:
        """
        Upload audio file to S3 and return the URL.
        """
        try:
            # Create audio-specific key
            timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
            key = f"audio/{timestamp}_{filename}_chunk_{chunk_index}.mp3"
            
            self.s3_client.put_object(
                Bucket=self.bucket_name,
                Key=key,
                Body=audio_data,
                ContentType='audio/mpeg'
            )
            
            # Return the S3 URL
            url = f"https://{self.bucket_name}.s3.amazonaws.com/{key}"
            logger.info(f"Successfully uploaded audio to S3: {url}")
            return url
            
        except ClientError as e:
            logger.error(f"Error uploading audio to S3: {str(e)}")
            raise ValueError(f"Failed to upload audio: {str(e)}")
    
    async def delete_file(self, key: str) -> bool:
        """
        Delete a file from S3.
        """
        try:
            self.s3_client.delete_object(
                Bucket=self.bucket_name,
                Key=key
            )
            logger.info(f"Successfully deleted file from S3: {key}")
            return True
            
        except ClientError as e:
            logger.error(f"Error deleting file from S3: {str(e)}")
            return False
    
    async def get_file_url(self, key: str) -> str:
        """
        Get the URL for a file in S3.
        """
        return f"https://{self.bucket_name}.s3.amazonaws.com/{key}"
    
    async def check_file_exists(self, key: str) -> bool:
        """
        Check if a file exists in S3.
        """
        try:
            self.s3_client.head_object(Bucket=self.bucket_name, Key=key)
            return True
        except ClientError:
            return False