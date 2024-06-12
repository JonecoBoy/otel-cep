# Executar
- executar algum dos executaveis ou via go run tempByCep.go
- Enviar alguma request para localhost:8080/temp/"cepCode"
- O retorno aparecerá no console e também na resposta.

# Executar com docker-compose
```shell
docker-compose up --build -d
```

para testar request pode-se usar os arquivos http dentro da inputApp



## Serviço A  -> Receber o CEP e validar string -> InputApp
porta 8091

## Serviço B  -> Temperatura Por CEP -> tempByCep
porta 8090


