# chirpstack-event-forward  

[中文](README_zh.md)

Forward chirpstack event data from chirpstack application http integration. Now we support pulsar and mqtt integration.


# Architecture

![Architecture](./docs/chirpstack-event-forward.png)

# Deployment

## Kubernetes

## Helm Chart  

Use chart from [chirpstack-event-forward](https://github.com/liangyuanpeng/charts/tree/main/chirpstack-event-forward).

require: helm v3

```
helm repo add lyp https://liangyuanpeng.github.io/charts
helm repo update
helm install lyp/chirpstack-event-forward
```

And then you can get the running pod of chirpstack-event-forward.

```shell
$ kubectl get po
NAME                                            READY   STATUS    RESTARTS   AGE
cef-chirpstack-event-forward-7fd74f9966-q5nzb   1/1     Running   0          6m22s
```