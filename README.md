# NSQ-connector
The NSQ connector connects OpenFaaS functions to NSQ topics, inspired by [Kafka-connector](https://github.com/openfaas-incubator/kafka-connector) project

# Deploy K8s Steps

1. First you must setup `faas` project in k8s

```
You can read docs from [OpenFaas](https://www.openfaas.com/) and setup faas
```

2. Deploy NSQ server && NSQ connector

```bash
kubectl create -f ./yaml/kubernetes
```

3. Deploy new function use faas-cli

```bash
# we use image feifeiiiiiiiiiii/openresty-openfaas-daemon-off
faas deploy --image feifeiiiiiiiiiii/openresty-openfaas-daemon-off --name=test-fn --annotation topic="nsq-fn-topic"
```

4. You can use nsqd rest api publish topic msg

``` bash
curl -H "Accept: application/json" -XPOST --data '{"topic": "nsq-fn-topic","value": "hello world"}' http://127.0.0.1:30151/pub?topic=faas-request
```

5. You can lookup pod nsq-connector logs as follow
```
2018/10/22 05:30:37 [#6] Received on [nsq-fn-topic]: 'hello world'
2018/10/22 05:30:37 Invoke function: test
2018/10/22 05:30:37 Response [200] from test Hello World
2018/10/22 05:30:38 Syncing topic map
2018/10/22 05:30:40 [#7] Received on [nsq-fn-topic]: 'hello world'
2018/10/22 05:30:40 Invoke function: test
2018/10/22 05:30:40 Response [200] from test Hello World
2018/10/22 05:30:41 Syncing topic map
2018/10/22 05:30:42 [#8] Received on [nsq-fn-topic]: 'hello world'
2018/10/22 05:30:42 Invoke function: test
2018/10/22 05:30:42 Response [200] from test Hello World
2018/10/22 05:30:44 Syncing topic map
2018/10/22 05:30:47 Syncing topic map
2018/10/22 05:30:50 Syncing topic map
```

6. Good luck have fun

## Configuration

This configuration can be set in the YAML files for Kubernetes or Swarm.

| env_var               | description                                                 |
| --------------------- |----------------------------------------------------------   |
| `upstream_timeout`      | Go duration - maximum timeout for upstream function call    |
| `rebuild_interval`      | Go duration - interval for rebuilding function to topic map |
| `topics`                | Topics to which the connector will bind                     |
| `gateway_url`           | The URL for the API gateway i.e. http://gateway:8080 or http://gateway.openfaas:8080 for Kubernetes       |
| `nsqlookupd`           | Default is `nsqlookupd`                                          |
| `print_response`        | Default is `true` - this will output the response of calling a function in the logs |