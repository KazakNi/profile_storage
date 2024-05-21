<<<<<<< HEAD
# profile_storage


Healthcheck example:

curl -u admin:admin http://localhost:8080/user/
 
curl -u admin:admin http://localhost:8080/user/2d4569f3-ed10-4ed9-8b7b-b5bcff7e1b56

POST. DELETE, PATCH методы рекомендуется отрабатывать через Postman где более удобно работать с cookies
=======
# Сервис по хранению профилей пользователей

Для разворачивания сервиса локально необходимо клонировать репозиторий:
```
git clone https://github.com/KazakNi/profile_storage.git
```

Перейти в директорию проекта

```
cd profile_storage
```
Заполнить файл конфига при необходимости в локации config/config.yaml

Запустить docker-compose

```
docker-compose up
```

Спецификация OpenAPI будет доступна по адресу http://localhost:8080/redoc

<img src="https://github.com/KazakNi/profile_storage/blob/main/swagger.jpg" > </img>

Healthcheck-команды через консоль:

```
curl -u admin:admin http://localhost:8080/user/
 ```
```
curl -u admin:admin http://localhost:8080/user/2d4569f3-ed10-4ed9-8b7b-b5bcff7e1b56
```
POST. DELETE, PATCH методы рекомендуется отрабатывать через Postman, где более удобно работать с cookies.

Юзернейм и пароль для тестирования - "admin:admin"

Тело для тестового POST-запроса:

```json
{
        "username": "aaaaaaa",
        "email": "example@test.ru",
        "password": "asdsdfsdf",
        "admin": true
}
```

Примеры отработки запросов:

<img src="https://github.com/KazakNi/profile_storage/blob/main/post.jpg" > </img>

<img src="https://github.com/KazakNi/profile_storage/blob/main/get.jpg" > </img>

Авторизация к ресурсам выполнена с помощью cookies, подписанных приватным ключом.


>>>>>>> e68fbedb8fafacba1f1ab5aa27bc17782d71a965

