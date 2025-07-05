.PHONY: dup rdup

dup:
	docker compose -f docker-compose.dev.yml up -d

rdup:
	docker compose -f ./docker-compose.dev.yml up --build -d

pup:
	docker compose -f docker-compose.dev.yml up -d

rpup:
	docker compose -f ./docker-compose.dev.yml up --build -d

ddown:
	docker compose -f docker-compose.dev.yml down

shut:
	docker stop $(shell docker ps -q)