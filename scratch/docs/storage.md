# Image Storage System

This document describes the image storage system for the nyumspace application, which provides flexible storage backend support for LocalStack S3 and AWS S3.

## Overview

The storage system provides:
- **Storage Abstraction**: Switch between LocalStack (development) and AWS S3 (production)
- **Image Upload**: Direct file upload and presigned URL generation
- **Image Management**: Automatic cleanup when homes are deleted
- **Security**: User authorization and content type validation

## Storage Backends

### LocalStack (Development)
- **Endpoint**: http://localhost:4566
- **Credentials**: test/test (default LocalStack credentials)
- **Bucket**: nyumspace-images (configurable)

### AWS S3 (Production)
- **Region**: Configurable (default: us-east-1)
- **Credentials**: AWS Access Key ID and Secret Access Key
- **Bucket**: Configurable

## Configuration

Set these environment variables:

```bash
# Development (LocalStack)
STORAGE_PROVIDER=localstack
S3_BUCKET=nyumspace-images

# Production (AWS S3)
STORAGE_PROVIDER=s3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
S3_BUCKET=your-production-bucket
```

## API Endpoints

### Direct Image Upload
```
POST /api/portal/homes/{home-id}/images/upload
Content-Type: multipart/form-data

Form field: image (file)
```

**Response:**
```json
{
  "image_url": "http://localhost:4566/nyumspace-images/images/homes/user123/home456/1640995200_uuid.jpg",
  "message": "Image uploaded successfully"
}
```

### Generate Presigned Upload URL
```
POST /api/portal/homes/{home-id}/images/presigned?content_type=image/jpeg
```

**Response:**
```json
{
  "upload_url": "http://localhost:4566/nyumspace-images/...",
  "image_key": "images/homes/user123/home456/1640995200_uuid.jpg",
  "message": "Presigned URL generated successfully"
}
```

### Delete Image
```
DELETE /api/portal/homes/{home-id}/images
Content-Type: application/json

{
  "image_url": "http://localhost:4566/nyumspace-images/images/homes/user123/home456/1640995200_uuid.jpg"
}
```

## Supported Image Types

- JPEG (image/jpeg)
- PNG (image/png)
- GIF (image/gif)
- WebP (image/webp)

## LocalStack Setup

1. Install LocalStack:
```bash
pip install localstack
```

2. Start LocalStack:
```bash
localstack start
```

3. Create the S3 bucket:
```bash
aws --endpoint-url=http://localhost:4566 s3 mb s3://nyumspace-images
```

4. Set bucket policy for public read access:
```bash
aws --endpoint-url=http://localhost:4566 s3api put-bucket-policy \
  --bucket nyumspace-images \
  --policy '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "PublicReadGetObject",
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::nyumspace-images/*"
      }
    ]
  }'
```

## Usage Examples

### Direct Upload with curl
```bash
curl -X POST \
  -H "Authorization: Bearer your-jwt-token" \
  -F "image=@/path/to/image.jpg" \
  http://localhost:3000/api/portal/homes/home123/images/upload
```

### Presigned URL Generation
```bash
curl -X POST \
  -H "Authorization: Bearer your-jwt-token" \
  "http://localhost:3000/api/portal/homes/home123/images/presigned?content_type=image/jpeg"
```

### Using Presigned URL (from frontend)
```javascript
// Get presigned URL from your API
const response = await fetch('/api/portal/homes/home123/images/presigned?content_type=image/jpeg', {
  method: 'POST',
  headers: { 'Authorization': 'Bearer ' + token }
});
const { upload_url } = await response.json();

// Upload directly to S3
await fetch(upload_url, {
  method: 'PUT',
  headers: { 'Content-Type': 'image/jpeg' },
  body: imageFile
});
```

## Implementation Details

### File Naming Convention
Images are stored with the following key structure:
```
images/homes/{user_id}/{home_id}/{timestamp}_{uuid}.{extension}
```

Example: `images/homes/user123/home456/1640995200_abc123.jpg`

### Automatic Cleanup
When a home is deleted, the associated image is automatically removed from storage.

### Security
- All endpoints require user authentication
- Users can only upload/delete images for homes they own
- Content type validation prevents non-image uploads
- File size limited to 10MB

## Development vs Production

The system automatically switches between LocalStack and AWS S3 based on the `STORAGE_PROVIDER` environment variable, making it easy to develop locally and deploy to production without code changes.
