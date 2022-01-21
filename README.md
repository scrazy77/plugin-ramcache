# Simple RAM Cache

This is a fork from https://github.com/traefik/plugin-simplecache

Simple cache plugin middleware caches responses on RAM.

## Configuration

To configure this plugin you should add its configuration to the Traefik dynamic configuration as explained [here](https://docs.traefik.io/getting-started/configuration-overview/#the-dynamic-configuration).
The following snippet shows how to configure this plugin with the File provider in TOML and YAML: 

Static:

```toml
[pilot]
  token="xxx"

[experimental.plugins.cache]
  modulename = "github.com/scrazy77/plugin-ramcache"
  version = "v0.1.1"
```

Dynamic:

```toml
[http.middlewares]
  [http.middlewares.my-cache.plugin.cache]
    maxExpiry = 300
```

```yaml
http:
  middlewares:
   my-cache:
      plugin:
        cache:
          maxExpiry: 300
```

### Options

#### Max Expiry (`maxExpiry`)

*Default: 300*

The maximum number of seconds a response can be cached for. The 
actual cache time will always be lower or equal to this.
	
#### Add Status Header (`addStatusHeader`)

*Default: true*

This determines if the cache status header `Cache-Status` will be added to the
response headers. This header can have the value `hit`, `miss` or `error`.

#### CacheQueryParams   (`cacheQueryParams`)

*Default: false*

This determines if query parameters will also be part of the cached response.

#### BlacklistedHeaders (`blacklistedHeaders`)

*Default: []*

This sets a lists of headers that will make the cache skip caching the request and response.

#### ForceNoCacheHeader (`forceNoCacheHeader`)

*Default: false*

This force sets a cache-control: no-cache headers when serve response.
