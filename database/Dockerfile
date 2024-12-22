FROM mariadb:10.9

COPY initialize.sql /docker-entrypoint-initdb.d/initialize.sql
COPY my.cnf /etc/mysql/conf.d/custom.cnf