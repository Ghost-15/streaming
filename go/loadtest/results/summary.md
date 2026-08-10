# k6 + runtime evidence

- Target: http://127.0.0.1:8080
- Scope: runtime metrics and pprof belong to this local instrumented target; it is not Render production.

| Palier | Source (kbit/s) | CPU p95 (% d'un coeur) | RSS max (MiB) | Drops | k6 checks | Requetes echouees | Duree HTTP p95 | Goroutines apres repos | Statut |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | :---: |
| 10 | 127.66 | 6.06 | 35.47 | 0 % | 100 % | 0 % | 83.282 s | 17 | OK |
| 100 | 128.73 | 10.08 | 42.75 | 0 % | 100 % | 0 % | 84.036 s | 17 | OK |
| 500 | 129.66 | 27.23 | 74.31 | 0 % | 100 % | 0 % | 83.703 s | 17 | OK |
