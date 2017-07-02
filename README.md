# go-byfahrer
Start a reverse-proxy to terminate SSL for those services, which do not deal with SSL themself.

## Pre-Production!

It's just a PoC hack for now, there are a couple of known limitation...

- [] proxy containers are not killed, when the service is killed they are hooked into
- [] error handling is not done propperly

## Example

Start the byfahrer (co-driver), which connects it to the docker API and allows it to subscribe to the events.

```bash
$ docker run -ti --rm -v /var/run/docker.sock:/var/run/docker.sock qnib/byfahrer
2017/07/02 11:57:18 [II] Start Version: 0.0.0
2017/07/02 11:57:18 [II] Dispatch broadcast for Back, Data and Tick
2017/07/02 11:57:18.857130 [NOTICE]   docker-events Name:docker-events >> Start docker-events collector v0.2.4
2017/07/02 11:57:18.857130 [NOTICE]     go-byfahrer Name:go-byfahrer >> Start plugin v0.0.0
2017/07/02 11:57:19.032062 [  INFO]     go-byfahrer Name:go-byfahrer >> Connected to 'moby' / v'17.06.0-ce-rc5'
2017/07/02 11:57:19.185778 [  INFO]   docker-events Name:docker-events >> Connected to 'moby' / v'17.06.0-ce-rc5'
```

If a container is started, which uses the label `org.qnib.byfahrer.proxy-image`, the agent will start a proxy agent.

```bash
$ docker run --rm --name www -ti -p 8081:8081 --label org.qnib.byfahrer.proxy-image=qnib/gosslterm qnib/plain-httpcheck
[II] qnib/init-plain script v0.4.28
> execute entrypoint '/opt/entry/00-logging.sh'
> execute entrypoint '/opt/entry/10-docker-secrets.env'
[II] No /run/secrets directory, skip step
> execute entrypoint '/opt/entry/99-remove-healthcheck-force.sh'
> execute CMD 'go-httpcheck'
2017/07/02 11:47:42 Start serving on 0.0.0.0:8080
```

**Please note**, that the http service serves on `8080`, but exposes `8081`. `8081` will be served by the proxy running in the same network namespace.

```bash
2017/07/02 12:01:58.486359 [  INFO]     go-byfahrer Name:go-byfahrer >> Use org.qnib.byfahrer.proxy-image=qnib/gosslterm to start proxy
2017/07/02 12:01:58.571868 [  INFO]     go-byfahrer Name:go-byfahrer >> Create proxy container 'www-proxy' for 'www'
```

Now one can query the proxy using SSL:

```bash
$ curl --insecure "https://127.0.0.1:8081/pi/99999" 
  Welcome: pi(99999)=3.141583
```

Which leads the http service to log the request...

```bash
request:+1|c app=go-httpcheck,endpoint=/pi,version=1.1.4
duration:239|ms app=go-httpcheck,endpoint=/pi,version=1.1.4
```

...and the proxy as well.

```bash
$ docker logs www-proxy                                                                                                                                                                             git:(master|…
2017/07/02 12:01:58 Load cert '/opt/qnib/ssl/cert.pem' and key '/opt/qnib/ssl/key.pem'
2017/07/02 12:01:58 Create http.Server on ':8081'
[negroni] 2017-07-02T12:02:24Z | 200 | 	 241.356539ms | 127.0.0.1:8081 | GET /pi/99999
```