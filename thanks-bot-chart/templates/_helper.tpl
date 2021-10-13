{{- define "deployment.app" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .service.deploymentName }}
spec:
  replicas: {{ .service.replicas | default 1 }}
  selector:
    matchLabels:
      app: {{ .service.appName }}
  template:
    metadata:
     labels:
       app: {{ .service.appName }}
    spec:
      containers:
        - image:  "d9nich/{{ .service.image.repository }}:{{ .service.image.tag }}"
          imagePullPolicy: {{ .pullPolicy }}
          name: {{ .service.appName }}
          ports:
            - containerPort: {{ .service.containerPort }} 
          resources:
            limits:
              memory: 256Mi
            requests:
              cpu: 75m
              memory: 64Mi
{{- end -}}


{{- define "service.app" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ .serviceName }}
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: {{ .containerPort }}
  selector:
    app: {{ .appName }}
{{- end -}}
