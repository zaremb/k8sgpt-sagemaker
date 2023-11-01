# Internal Docs

## Prerequisites

Install go 1.20
Install kind

```bash
go env -w GOPRIVATE="*"
go env -w GOPROXY="direct"
```

## Prepare cluster for analyze

```bash
kind create cluster
```

Check your current context

```bash
kubectl config current-context
//Should be kind-kind
```

Deploy invalid pod

```bash
kubectl run nginx --image=nginxx --restart=Never
```

Check created pod
```
kubectl get pods
NAME    READY   STATUS             RESTARTS   AGE
nginx   0/1     ImagePullBackOff   0          123m
```

## Build binary from code

in main directory

```bash
Change endpoint in pkg/ai/sagemaker.go
go build .
```

## First config of AmazonSageMaker Provider

```bash
./k8sgpt auth list
./k8sgpt auth add --backend amazonsagemaker
  //  type random string as a key
 
  // Make amazonsagemaker as a default provider
./k8sgpt auth default -p amazonsagemaker 
```

## Test amazonsagemaker provider

```bash
./k8sgpt analyze --explain
```
