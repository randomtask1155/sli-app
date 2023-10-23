

# push app

first push app with no start to get app guid

update env vars in manifest.yml except for CF_PING_INSTANCE.  We need to push the app before we can fill in that variable.

```
GOROUTER_LIST: 10.1.1.119:10.1.1.66
CF_PING_INSTANCE: APP-GUID:1
CF_APP_DOMAIN: sli-app.apps.domain.com
PING_SLEEP_INTERVAL_SECONDS: "2"
```


```
cf push -i 2 --no-start
```

get app guid

```
cf app sli-app --guid
86b0587c-a87c-40e1-91b7-e83b82b51d56
```

update CF_PING_INSTANCE env var in manifest.yml.  This will cause app index 0 to ping index 1 across all the gorouters. App index 1 will not ping it self and remain idle waiting for requests. 

```
CF_PING_INSTANCE: 86b0587c-a87c-40e1-91b7-e83b82b51d56:1
```

Push the app again to start it

```
cf push -i 2
```


