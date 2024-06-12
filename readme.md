# Executar
- executar algum dos executaveis ou via go run tempByCep.go
- Enviar alguma request para localhost:8080/temp/"cepCode"
- O retorno aparecerá no console e também na resposta.

# Executar com docker-compose
```shell
docker-compose up --build -d
```

# CloudRun
- o endereço que está online da cloudRun é : https://temp-cep-7h5a3mqufa-uc.a.run.app
- testar https://temp-cep-7h5a3mqufa-uc.a.run.app/temp/"cepCode"

# Teste Automatizado

## Teste da api / integracao
```shell
go test -v ./pkg/tempByCep.go ./pkg/tempByCep_test.go
```

## Teste individuais

Testes apis e metodos de cep
```shell
go test -v ./pkg/external/brasilApiCep.go ./pkg/external/viaCEP.go ./pkg/external/cepAPIs_test.go
```
Testes de metodos de temperatura
```shell
go test -v ./pkg/external/weatherApi.go ./pkg/external/weatherApi_test.go
```
