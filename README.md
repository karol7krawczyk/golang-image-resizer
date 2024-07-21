# Golang Image Resizer Service

![License](https://img.shields.io/badge/license-MIT-blue.svg)


## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration File](#configuration-file)
- [Nginx Configuration](#nginx-configuration)
- [Testing](#testing)
- [License](#license)

## Overview
The Image Resizing Service is a web server for dynamic image resizing. It allows image resizing based on specified dimensions through HTTP requests. The service supports multiple image formats and can be configured to serve images from various directories and routes.

## Features
- **Dynamic Resizing:** Adjust image dimensions using query parameters or shorthand notation.
- **Format Support:** Handles JPEG, PNG, and WebP formats.
- **Configurable:** Manage routes and directories via an INI file.
- **Integration with Nginx:** Supports caching and request proxying for improved performance.

## Features
- Resize images to specified dimensions using query parameters.
- Supports multiple image formats: JPEG, PNG, and WebP.
- Configurable through an ini file.
- Integrates with Nginx for caching and proxying requests.


## Installation
To get started, clone the repository and install the required dependencies.


### Clone the repository
```bash
git clone https://github.com/Karol7Krawczyk/golang-resize-image.git
cd golang-resize-image
```

### Run in docker
```bash
make build
make up
```

### The binary file is available after building the docker
```bash
./resizer
```

## Usage
To resize an image, send a GET request with the desired dimensions.Using the nginx configuration, we can set caching parameters and rate limit. Here's how you can use the resizer:

- By Query Parameters: http://localhost/images/test.jpg?width=300&height=500
- By Shortened Parameters: http://localhost/images/test.jpg?300x500
- Original Image in Another Format: http://localhost/images/test.webp


## Configuration File (config.ini)
Define the routes and directories for serving images by creating a config.ini file:

```bash
  [server]
  port = 8080

  [route_images]
  route = /images/
  dir = ./public/images

  [route_extra_images]
  route = /extra_images/
  dir = ./public/extra_images
```

## Nginx Configuration  (nginx.conf)
Below is a sample Nginx configuration to set up caching and proxying:

```bash
worker_processes 1;

events { worker_connections 1024; }

http {
    include mime.types;
    default_type application/octet-stream;
    keepalive_timeout 65;
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    types_hash_max_size 2048;    

    proxy_cache_path /tmp/nginx_cache levels=1:2 keys_zone=my_cache:10m inactive=24h max_size=1g;
    limit_req_zone $binary_remote_addr zone=my_resource:10m rate=5r/s; 

    server {
        listen 80;
        server_name yourdomain.com;

        location /images/ {
            limit_req zone=my_resource burst=10 nodelay;

            proxy_pass http://golang:8080/images/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            proxy_cache my_cache;
            proxy_cache_valid 200 24h;
            proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
            add_header X-Cache-Status $upstream_cache_status;
        }

        location /extra_images/ {
            limit_req zone=my_resource burst=10 nodelay;

            proxy_pass http://golang:8080/extra_images/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            proxy_cache my_cache;
            proxy_cache_valid 200 24h;
            proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
            add_header X-Cache-Status $upstream_cache_status;
        }

        location / {
            root /app/public;
            try_files $uri $uri/ =404;
        }
    }  
}

```


## Testing
To run the tests, use the following command after build docker-compose:

```bash
make up
make test
```

## License
This project is licensed under the MIT License - see the LICENSE file for details.
