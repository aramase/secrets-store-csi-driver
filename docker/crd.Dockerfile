FROM alpine as builder
ARG KUBE_VERSION=v1.21.2
ARG TARGETARCH
ARG TARGETPLATFORM
ARG TARGETOS

RUN apk add --no-cache curl && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$KUBE_VERSION/bin/linux/$TARGETARCH/kubectl && \
    chmod +x kubectl

FROM scratch
COPY * /crds/
COPY --from=builder /kubectl /kubectl
ENTRYPOINT ["/kubectl"]
