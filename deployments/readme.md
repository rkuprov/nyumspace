1) To spin up the DB:
run in the 'deployments' folder:
```bash
  docker-compose up -d
```
Bring it down with:
```bash
  docker-compose down
```
in the same directory.

2) Run Migrations
```bash
  goose migrate {up/down}
```

``` bash
  river migrate-up --line main
```