# Supernova Calc — REST API

## Методы API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/telescopes?min_aperture=100 | Список опубликованных телескопов с фильтром по диаметру |
| POST | /api/telescopes | Создать черновик с картинкой и видео |
| GET | /api/draft | Получить черновик текущего пользователя |
| PUT | /api/telescopes/:id/publish | Опубликовать услугу |
| GET | /api/feed | Лента (первый опубликованный) |
| GET | /api/feed?id=1&next=true | Лента — следующий по ID |
| POST | /api/telescopes/:id/like | Поставить (like=1) или снять (like=0) лайк |
| DELETE | /api/telescopes/:id | Логическое удаление услуги |
| POST | /api/users/register | Регистрация пользователя |
| POST | /api/users/login | Заглушка авторизации |
| POST | /api/users/logout | Заглушка деавторизации |

## Таблицы БД

- `users` — пользователи
- `telescopes` — услуги
- `likes` — лайки (многие-ко-многим)
