down_mysql:
	docker-compose -f docker-compose-mysql.yml down

down_postgres:
	docker-compose -f docker-compose-postgres.yml down

up_mysql:
	echo "starting authentication service with mysql backend database"
	docker-compose -f docker-compose-mysql.yml up --remove-orphans

up_postgres:
	echo "starting authentication service with postgres backend database"
	docker-compose -f docker-compose-postgres.yml up --remove-orphans
