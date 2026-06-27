DOCKER_USER  ?= srvsurya
IMAGE_NAME   ?= system-monitor
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
FULL_IMAGE   = $(DOCKER_USER)/$(IMAGE_NAME):$(VERSION)
LATEST_IMAGE = $(DOCKER_USER)/$(IMAGE_NAME):latest

.PHONY: help build up down logs push tag clean

help:
	@echo ""
	@echo "  make build     Build the Docker image"
	@echo "  make up        Start all containers (detached)"
	@echo "  make down      Stop and remove containers"
	@echo "  make logs      Tail app logs"
	@echo "  make tag       Tag image for Docker Hub"
	@echo "  make push      Push to Docker Hub"
	@echo "  make clean     Remove containers, volumes, and image"
	@echo ""

## Build image via Docker Compose
build:
	docker compose build --no-cache

## Start containers
up:
	docker compose up -d
	@echo "→ Running at http://localhost:80"

## Stop containers
down:
	docker compose down

## Tail logs
logs:
	docker compose logs -f app

## Tag the built image for Docker Hub
tag:
	docker tag system-monitor:latest $(FULL_IMAGE)
	docker tag system-monitor:latest $(LATEST_IMAGE)
	@echo "Tagged: $(FULL_IMAGE)"
	@echo "Tagged: $(LATEST_IMAGE)"

## Push to Docker Hub (must be logged in: docker login)
push: tag
	docker push $(FULL_IMAGE)
	docker push $(LATEST_IMAGE)
	@echo "Pushed to Docker Hub"

## Full clean including named volumes (WARNING: deletes SQLite data)
clean:
	docker compose down -v --rmi local