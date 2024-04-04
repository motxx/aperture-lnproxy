all: up
clean:
	docker compose down
	rm -rf ./db/logs/ && rm -rf ./db/postgres/data/
up:
	docker compose up --build -d
logs:
	docker compose logs -f
prune:
	docker system prune
build-cli:
	cd ./contents && make
