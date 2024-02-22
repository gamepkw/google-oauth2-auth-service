# # Define variables
# DOCKER := docker
# DOCKER_COMPOSE := docker-compose

# # Define targets and dependencies
# .PHONY: build run stop clean

# # Build Docker image
# build:
# 	$(DOCKER_COMPOSE) build

# # Run Docker container
# run:
# 	$(DOCKER_COMPOSE) up -d

# # Stop Docker container
# stop:
# 	$(DOCKER_COMPOSE) down

# # Clean up
# clean:
# 	$(DOCKER_COMPOSE) down --volumes --remove-orphans

# # Default target
# .DEFAULT_GOAL := build-run

# # Build and run Docker container by default
# build-run: build run

build:
	docker build -t google-oauth2-auth-service:latest .

tag:
	docker tag google-oauth2-auth-service:latest docker.io/gamepkw/google-oauth2-auth-service:latest

push:
	docker push gamepkw/google-oauth2-auth-service:latest

stop:
	docker stop google-oauth2-auth-service-container || true

remove:
	docker rm google-oauth2-auth-service-container || true

run:
	docker run -d -p 9090:9090 --name google-oauth2-auth-service-container google-oauth2-auth-service:latest


#make build && make tag && make push && make stop && make remove && make run
#make build && make stop && make remove && make run
