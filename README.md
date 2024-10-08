## Описание
> [!NOTE]
> Проект написан в рамках курса "Продвинутый Go-разработчик" от [Яндекс.Практикум](https://practicum.yandex.ru).
> C техническим заданием можно ознакомиться в файле [SPECIFICATION.md](SPECIFICATION.md)
>
> Для защиты проекта была подготовлена презентация и текст, их можно найти на [Google Drive](https://drive.google.com/drive/folders/1PzV5gYQheW_rV_HHJRlFxnsO637FmI5B?usp=sharing).

**gophermart-loyalty-service** - сервис отвечающий за систему лояльности маркетплейса gophermart,
а именно за регистрацию и авторизацию пользователей в системе, добавление новых заказов к обработке, 
получение расчитанных баллов, начисление и списание бонусов и вывод текущего баланса пользователя.
_(за расчет количества начисленных баллов отвечает отдельный сервис gophermart-accrual-service)_.

## Makefile
Для упрощения локальной разработки БД, миграции и сам сервис запускаются в docker контейнерах.
Некоторые основные команды для работы с ними вынесены в Makefile:

| Команда    | Описание                              |
|------------|---------------------------------------|
| build      | Собрать образы                        |
| up         | Поднять контейнеры                    |
| down       | Убить контейнеры                      |
| logs       | Вывести логи                          |
| logsf      | Вывести и продолжать выводить логи    |
| migrate    | Применить миграции                    |
| diff       | Сгенерировать миграцию                |
| vet        | Запустить go vet                      |
| test       | Запустить go test                     |
| test-race  | Запустить go test -race               |
| test-cover | Вычислить процент покрытия тестами    |
| pprof-cpu  | Записать профиль использования CPU    |
| pprof-mem  | Записать профиль использования памяти |

## Кофигурация
Сервис поддерживает конфигурацию через переменные окружения и флаги процесса. В приоритете параметры переданные через флаг.

| Переменная окружения   | Флаг                          | Описание                                                                        | По-умолчанию   |
|------------------------|-------------------------------|---------------------------------------------------------------------------------|----------------|
| APP_ENV                | -e / --env                    | Текущая среда приложения                                                        | dev            |
| APP_SECRET             | -s / --app-secret             | Ключ шифрования JWT                                                             | aPp$eCr3t      |
| RUN_ADDRESS            | -a / --address                | Адрес приложения                                                                | :8080          |
| ACCRUAL_SYSTEM_ADDRESS | -r / --accrual-system-address | Адрес gophermart-accrual-service                                                | localhost:8081 |
| DATABASE_URI           | -d / --database-uri           | URI базы данных                                                                 |                |
| RETRIEVER_CONCURRENCY  | --retriever-concurrency       | Максимальное кол-во горутин получающих статус расчета и расчитанные баллы       | 10             |
| ROUTER_CONCURRENCY     | --router-concurrency          | Максимальное кол-во горутин распределяющих полученные ответы по статус-очередям | 10             |
| PROCESSING_CONCURRENCY | --processing-concurrency      | Максимальное кол-во горутин обрабатывающих поступления в статусе PROCESSING     | 10             |
| INVALID_CONCURRENCY    | --invalid-concurrency         | Максимальное кол-во горутин обрабатывающих поступления в статусе INVALID        | 10             |
| PROCESSED_CONCURRENCY  | --processed-concurrency       | Максимальное кол-во горутин обрабатывающих поступления в статусе PROCESSED      | 10             |
| UPDATE_BATCH_SIZE      | --update-batch-size           | Максимальное кол-во заказов обрабатываемых одной горутиной                      | 100            |
| LOG_LEVEL              | -l / --log-level              | Уровень логирования                                                             | info           |
| CPU_PROFILE_FILE       | --cpu-profile-file            | Файл для записи профиля использования CPU                                       | ./cpu.pprof    |
| CPU_PROFILE_DURATION   | --cpu-profile-duration        | Время записи профиля использования CPU                                          | 30s            |
| MEM_PROFILE_FILE       | --mem-profile-file            | Файл для записи профиля использования памяти                                    | ./mem.pprof    |
| SHUTDOWN_TIMEOUT       | --shutdown-timeout            | Время отведенное на нормальное завершение внутренних процессов приложения       | 15s            |

## Структура проекта

| Директория | Субдиректория | Содержимое                                                                                                                                                                                                                              |
|-----------:|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|    .docker |               |                                                                                                                                                                                                                                         |
|          - | atlas         | Dockerfile для [atlas](https://atlasgo.io/)                                                                                                                                                                                             |
|          - | service       | Dockerfile для gophermart-loyalty-service                                                                                                                                                                                               |
|    .github |               | CI/CD конфиги                                                                                                                                                                                                                           |
|        cmd |               |                                                                                                                                                                                                                                         |
|          - | accrual       | Бинарники mock-сервиса accrual                                                                                                                                                                                                          |
|          - | gophermart    | Код входной точки в основное приложение                                                                                                                                                                                                 |
|   internal |               | Внутренние пакеты приложения                                                                                                                                                                                                            |
|          - | accrual       | Клиент для общения с сервисом accrual                                                                                                                                                                                                   |
|          - | app           | DI и запуск/остановка основных горутин приложения                                                                                                                                                                                       |
|          - | config        | Обработка переменных окружения и флагов процесса                                                                                                                                                                                        |
|          - | context       | Абстракция для передачи ID пользователя через контекст                                                                                                                                                                                  |
|          - | controller    | Хендлеры HTTP-запросов, работа с JSON                                                                                                                                                                                                   |
|          - | entity        | Сущности, хранимые в БД                                                                                                                                                                                                                 |
|          - | jwt           | Работа с JWT                                                                                                                                                                                                                            |
|          - | logger        | Логирование                                                                                                                                                                                                                             |
|          - | manager       | Фасады для работы с репозиториями                                                                                                                                                                                                       |
|          - | middleware    | HTTP-Middleware (аутентификация, recover)                                                                                                                                                                                               | 
|          - | processor     | Обработчики добавленных пользователем заказов                                                                                                                                                                                           |
|          - | repository    | Репозитории БД                                                                                                                                                                                                                          |
|          - | router        | Конфигурирование endpointов, прокидывание middleware                                                                                                                                                                                    |
|          - | server        | Конфигурирование HTTP-сервера                                                                                                                                                                                                           |
| migrations |               | Миграции БД                                                                                                                                                                                                                             |
|        pkg |               | Доступные к переиспользованию пакеты                                                                                                                                                                                                    |
|          - | client        | Go-клиент для HTTP-интерфейса приложения                                                                                                                                                                                                |
|          - | generator     | Реализация паттерна генератор                                                                                                                                                                                                           |
|          - | gorm          | Расширения для [gorm](https://gorm.io/) (типы bcrypt, money)                                                                                                                                                                            |
|          - | http          | Расширения для http (обработчик заголовка Retry-After)                                                                                                                                                                                  |
|          - | middleware    | HTTP-Middleware (комрессия, декомпрессия, интеграция с [zap](https://github.com/uber-go/zap))                                                                                                                                           |
|          - | pprof         | Фасад для записи профилей pprof                                                                                                                                                                                                         |
|          - | queue         | Реализация структуры очередь                                                                                                                                                                                                            |
|          - | requests      | Модели запросов к сервису                                                                                                                                                                                                               |
|          - | responses     | Модели ответов сервиса                                                                                                                                                                                                                  |
|          - | retry         | Реализация retry логики                                                                                                                                                                                                                 |
|          - | semaphore     | Реализация примитива синхронизации семафор                                                                                                                                                                                              |
|          - | validator     | Валидация данных ([алгоритм Луна](https://ru.wikipedia.org/wiki/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC_%D0%9B%D1%83%D0%BD%D0%B0), положительное число), интеграция с [govalidator](https://github.com/asaskevich/govalidator) |

## Используемые сторонние пакеты

| Пакет                                                                                             | Описание                       |
|---------------------------------------------------------------------------------------------------|--------------------------------|
| [caarlos0/env](https://github.com/caarlos0/env)                                                   | Обработка переменных окружения |
| [spf13/pflag](https://github.com/spf13/pflag)                                                     | Обработка флагов процесса      |
| [uber-go/zap](https://github.com/uber-go/zap)                                                     | Логирование                    |
| [go-chi/chi](https://github.com/go-chi/chi)                                                       | HTTP-роутинг                   |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt)                                               | Работа с JWT                   |
| [asaskevich/govalidator](https://github.com/asaskevich/govalidator)                               | Валидация данных               |
| [go-resty/resty](https://github.com/go-resty/resty)                                               | HTTP-клиенты                   |
| [go-gorm/gorm](https://github.com/go-gorm/gorm)                                                   | ORM, DBAL                      |
| [ariga/atlas](https://github.com/ariga/atlas) и [pressly/goose](https://github.com/pressly/goose) | Миграции БД                    |
| [stretchr/testify](https://github.com/stretchr/testify)                                           | Unit-тестирование              |
| [ovechkin-dm/mockio](https://github.com/ovechkin-dm/mockio)                                       | Создание mockов "на лету"      |
| [data-dog/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)                                     | Mockи SQL запросов             |
