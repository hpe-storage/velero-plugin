FROM debian:stretch-slim
RUN mkdir /plugins
ADD velero-* /plugins/
USER nobody:nobody
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
