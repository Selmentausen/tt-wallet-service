build:
	docker-compose build

run:
	docker-compose up

test:
	go test ./...

stop:
	docker-compose down

clean:
	docker-compose down -v