# HTTP Firewalled-Proxy

This is a simple HTTP proxy with firewall.
It uses a backend URL database to filter out requests. 
The database is accessed via HTTP GET. It can be sharded to distribute loads across multiple backends.
The proxy front end lookus up the URL database upon HTTP requests from clients and if the URL is blacklisted,
it will drop the request and return HTTP 403, otherwise it will act as an HTTP proxy.


TODO: 

- Use bolt/etcd DB to store blacklisted URLs.
- Use consistent hashing to select backends.
- Create and use Kubernetes containers for the proxy and the backend.
- Add Minikube tests using Kubernetes containers.
- Enable loadbalancing at the front end.
- Use multiple backends to shard URLs.
 
