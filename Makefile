#!make

run:
	killport 8080
	air -- -log error

# Generate tailwindcss classes for development
tw:
	tailwindcss -i tailwind.base.css -o web/static/tailwind.css --watch

# Docker Build
docker-build:
	tailwindcss -i tailwind.base.css -o web/static/tailwind.css --minify
	docker build -t pengoe .

# Docker Run
docker-run:
	docker run -p 8080:8080 --env-file .env pengoe

# Push migration
push:
	turso db shell $(db) < internal/db/schema.sql
