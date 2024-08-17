# Задание
Написать часть сервиса аутентификации.

## Два REST маршрута
 - Первый маршрут выдает пару Access, Refresh токенов для пользователя с идентификатором (GUID) указанным в параметре запроса
 - Второй маршрут выполняет Refresh операцию на пару Access, Refresh токенов

## Используемые технологии
 - Go
 - JWT
 - PostgreSQL

## Требования
Access токен тип JWT, алгоритм SHA512, хранить в базе строго запрещено.

Refresh токен тип произвольный, формат передачи base64, хранится в базе исключительно в виде bcrypt хеша, должен быть защищен от изменения на стороне клиента и попыток повторного использования.

Access, Refresh токены обоюдно связаны, Refresh операцию для Access токена можно выполнить только тем Refresh токеном который был выдан вместе с ним.

Payload токенов должен содержать сведения об ip адресе клиента, которому он был выдан. В случае, если ip адрес изменился, при рефреш операции нужно послать email warning на почту юзера (для упрощения можно использовать моковые данные).

## Результат

Результат выполнения задания нужно предоставить в виде исходного кода на Github. Будет плюсом, если получится использовать Docker и покрыть код тестами.

P.S. Друзья! Задания, выполненные полностью или частично с использованием chatGPT видно сразу. Если вы не готовы самостоятельно решать это тестовое задание, то пожалуйста, давайте будем ценить время друг друга и даже не будем пытаться :)

# Запуск
```make run_docker```

# Обзор сервиса
 - Swagger: http://localhost:8080/swagger/index.html
 - SMTP: http://localhost:8025


// TODO: добавить ссылки на swagger, mailhog, и сам сервис

# Пометки по заданию
1. Нужно было бы сделать конфиг который пробрасывает всему приложение в каком режиме запускаться.
2. Считаю следовало бы оставить больше комментариев в тестах для обьяснения что просиходит