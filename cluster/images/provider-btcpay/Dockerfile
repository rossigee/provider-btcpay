FROM gcr.io/distroless/static:nonroot

# OpenContainers Image Spec Labels
LABEL org.opencontainers.image.title="provider-btcpay"
LABEL org.opencontainers.image.description="Crossplane provider for BTCPay Server (stores, invoices, webhooks)"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-btcpay"
LABEL org.opencontainers.image.url="https://github.com/rossigee/provider-btcpay"
LABEL org.opencontainers.image.documentation="https://github.com/rossigee/provider-btcpay/blob/master/README.md"

ARG TARGETOS
ARG TARGETARCH

ADD bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/provider

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/provider"]