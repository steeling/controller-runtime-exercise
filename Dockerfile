FROM gcr.io/distroless/base
COPY dist/my-app-controller /usr/local/bin/my-app-controller

ENTRYPOINT [ "usr/local/bin/my-app-controller" ]