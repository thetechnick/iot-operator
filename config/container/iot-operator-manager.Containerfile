FROM scratch

WORKDIR /
COPY passwd /etc/passwd
COPY iot-operator-manager /

USER "noroot"

ENTRYPOINT ["/iot-operator-manager"]
