## Monitoring & Observability

Docker-compose ships with integrated Prometheus and Grafana:
- `/metrics` endpoint on port 9090 is scraped every 5 seconds
- Grafana dashboard available at [http://localhost:3000](http://localhost:3000) (default login: admin/admin)
- Add `Prometheus` datasource pointing to `http://prometheus:9090`
- Import sample dashboard from `grafana/tcp-lb-dashboard.json` (приложи его в репозиторий)

Default panels:
- Active connections by backend
- Backend health status
- Connection errors
- Connection duration histogram
