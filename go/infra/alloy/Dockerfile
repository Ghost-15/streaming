FROM grafana/alloy:v1.18.1

COPY go/infra/alloy/config.alloy /etc/alloy/config.alloy

ENTRYPOINT ["/bin/alloy"]
CMD ["run", "--server.http.listen-addr=0.0.0.0:12345", "/etc/alloy/config.alloy"]
