.PHONY: run tailwind-watch templ-watch go-watch tailwind-install

run:
	@bash -c ' \
		trap "echo Stopping...; kill 0" SIGINT; \
		$(MAKE) tailwind-watch & \
		$(MAKE) templ-watch & \
		$(MAKE) sqlc-watch & \
		$(MAKE) go-watch & \
		wait'

tailwind-watch:
	./tmp/tailwindcss -i assets/input.css -o assets/css/output.css --watch 

templ-watch:
	@templ generate --watch

sqlc-watch:
	@watchexec -w db/migrations -w db/queries -e sql sqlc generate

sqlc:
	@sqlc generate

go-watch:
	wgo -file=.css -file=.js -file=.go go run ./cmd/main.go

tailwind-install:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
	chmod +x tailwindcss-macos-arm64
	mv tailwindcss-macos-arm64 ./tmp/tailwindcss

pg:
	docker run --name explorer-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=explorer -d -p 5444:5432 postgres
