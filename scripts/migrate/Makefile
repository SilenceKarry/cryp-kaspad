MYSQL_URL="mysql://root:123456@tcp(127.0.0.1:3306)/eos?charset=utf8&parseTime=True&loc=Local"
MYSQL_SQL_PATH="migrations/mysql"

db-mysql-init:
	migrate create -ext sql -dir ${MYSQL_SQL_PATH} -seq init_schema

db-mysql-up:
	migrate --path ${MYSQL_SQL_PATH} --database ${MYSQL_URL} up

db-mysql-down:
	migrate --path ${MYSQL_SQL_PATH} --database ${MYSQL_URL} down

