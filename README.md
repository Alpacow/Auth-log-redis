# Auth-log-redis
Basic parse auth.log with golang and redis

Inicializar docker:
	systemctl start docker

Instalação:
	sudo docker-compose build

Atualizar após modificação:
	sudo docker-compose up

Rodar:
	sudo docker run -p 6060:6379 --name logredis  --rm logredis
