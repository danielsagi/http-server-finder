# http-server-finder
Minified golang multithreaded http server finder by response headers regex

# Usage Examples
1. finding CONNECT enabled http servers 
```sh
./http-server-finder -t 5 -n 200 -X OPTIONS -k Allow -r CONNECT -w target_urls.txt -o output.txt
```
