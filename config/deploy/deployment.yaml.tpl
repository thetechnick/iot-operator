apiVersion: apps/v1
kind: Deployment
metadata:
  name: iot-operator
  namespace: iot-system
  labels:
    app.kubernetes.io/name: iot-operator
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: iot-operator
  template:
    metadata:
      labels:
        app.kubernetes.io/name: iot-operator
    spec:
      serviceAccountName: iot-operator
      containers:
      - name: manager
        image: quay.io/nico_schieder/iot-operator-manager:latest
        args:
        - --enable-leader-election
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            cpu: 100m
            memory: 50Mi
          requests:
            cpu: 100m
            memory: 50Mi
