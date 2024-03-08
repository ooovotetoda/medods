# JWT Auth REST API

Данный проект - RESTful API для авторизации пользовавтеля по JWT токену.


## Запуск проекта

Для запуска необходимо ввести команду
```
go run cmd/medods/main.go
```

## Технологии
```Go``` 
```MongoDB``` 
```JWT```

## Получение пары Access & Refresh токенов

Для получения пары Access и Refresh токенов необходимо выполнить POST-запрос по адресу `/auth` с JSON-телом запроса, содержащим `guid` пользователя:
```
POST /auth

{
  "guid": "<guid>"
}
```



## Обновление пары Access & Refresh токенов

Для обновления пары Access и Refresh токенов необходимо выполнить POST-запрос по адресу `/auth/refresh` с JSON-телом запроса, содержащим `refresh_token`:


```
POST /auth/refresh

{
  "refresh_token": "<refresh_token>"
}
```


