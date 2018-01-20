# goHttpServer
Slightly less simple than the simpleHttpServer with added https redirect and logging capabilities.
Supports logging directly to JSON.


```
  Usage of ./goHttpServer:
  -p string
    Port to listen on. Kinda optional, will use 80 if not provided
  -c string
    (optional) -c Path to cert chain
  -d string
    (optional) -d Path to directory to serve
  -k string
    (optional) -k Path to cert private key
  -l string
    (optional) -l Log file to write access logs
  -j	
    (optional) -j Saves log results as JSON. Requires logfile to be provided
  -r
    (optional) -r Redirect using a web server on port 80 to redirect to port 443
``` 
