# Supernova Calc — REST API

## Методы API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/telescopes?min_aperture={int}&is_mine={bool} | Список опубликованных телескопов с фильтром по диаметру и автору |
| GET | /api/feed?id={int}&next={bool} | Лента: без параметров — первый; с id — конкретный; с next=true — следующий |
| GET | /api/draft | Черновик текущего пользователя (не более 1) |
| POST | /api/telescopes | Создать черновик + загрузить картинку и видео |
| PUT | /api/telescopes/{id}/publish | Опубликовать услугу (статус - published) |
| DELETE | /api/telescopes/{id} | Логическое удаление (soft delete) |
| POST | /api/telescopes/{id}/like | Поставить (like=1) или снять (like=0) лайк |
| POST | /api/users/register | Регистрация пользователя |
| POST | /api/users/login | Аутентификация |
| POST | /api/users/logout | Деавторизация |

## Таблицы БД

- `users` — пользователи
- `telescopes` — услуги
- `likes` — лайки (многие-ко-многим)
