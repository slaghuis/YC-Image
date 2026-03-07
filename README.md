# YB-Image
Image management micro-service for the Yellow Bus company

For image resizing

go get github.com/h2non/bimg
 and 
Make sure libvips is installed on your system (Linux recommended).
For Ubuntu:
sudo apt-get install -y libvips-dev


** Resize Strategy **
When a user uploads an image:
Save original
Create main resized image
- product: max 1200px
- banner: max 1920px
- blog: max 1000px
- user: max 512px

Create thumbnail
- default: 300x300 (square crop)



NGINX
server {
    listen 80;
    server_name images.mydomain.com;

    # Root folder for all image categories
    root /var/www;

    # Serve images at /images/* directly from filesystem
    location /images/ {
        alias /var/www/images/;
        autoindex off;

        # Add caching headers
        expires 30d;
        add_header Cache-Control "public, max-age=2592000, immutable";

        # Security headers
        add_header X-Content-Type-Options nosniff;
    }

    # Optional: restrict access to original images
    location ~* /images/.+-orig\.(jpg|jpeg|png|webp)$ {
        deny all;
    }

    # Fallback
    location / {
        return 404;
    }
}
