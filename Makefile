all: daemon
clean:
	docker compose down
	rm -rf ./db/logs/ && rm -rf ./db/postgres/ && rm -rf ./db/
daemon:
	docker compose up --build -d
build-cli:
	cd ./contents && make