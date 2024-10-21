build:
	@go build -o entaingo

run:
	@go run .

test:
	@go test -v ./...

testCover:
	@go test -v ./... -cover

swagger:
	@$(HOME)/go/bin/swag init -g ./src/routes/routers.go

dockerize:
	@docker build -t entaingo:latest .

dockerrun:
	@docker run --name entaingo -p 4000:4000 entaingo:latest

dockerCompose:
	@docker-compose up --build --force-recreate

