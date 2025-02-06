# Постановка задачи

Тема: поиск данных по XML и тестовое покрытие

Это комбинированное задание по тому, как отправлять запросы, получать ответы, работать с параметрами, хедерами, а так же писать тесты.

У нас есть какой-то поисковый сервис:
* SearchClient - структура с методом FindUsers, который отправляет запрос во внешнюю систему и возвращает результат, немного преобразуя его. Находится в файле client.go, править нельзя.
* SearchServer - своего рода внешняя система. Непосредственно занимается поиском данных в файле `dataset.xml`. В продакшене бы запускалась в виде отдельного веб-сервиса, но в вашем коде запускается как отдельный хендлер.

Требуется:
* Написать функцию SearchServer в файле `client_test.go`, который вы будете запускать в тесте через тестовый сервер (`httptest.NewServer`, пример использования в `4/http/server_test.go`)
* Покрыть тестами метод FindUsers, чтобы покрытие файла `client.go` было максимально возможным. Тесты писать в `client_test.go`.
* Так же требуется сгенерировать html-отчет с покрытием.
* Тесты надо писать полноценные, те не чтобы получить покрытие, а которые реально тестируют ваш код, проверяют возвращаемый результат, граничные случаи и тд. Они должны показывать что SearchServer работает правильно.
* Из предыдущего пункта вытекает что SearchServer тоже надо писать полноценный

SearchServer принимает GET-параметры:
* `query` - что искать. Ищем по полям записи `Name` и `About` просто подстроку, без регулярок. `Name` - это first_name + last_name из xml. Если поле пустое - то возвращаем все записи.
* `order_field` - по какому полю сортировать. Работает по полям `Id`, `Age`, `Name`, если пустой - то сортируем по `Name`, если что-то другое - SearchServer ругается ошибкой. 
* `order_by` - направление сортировки (как есть, по убыванию, по возрастанию), в client.go есть соответствующие константы
* `limit` - сколько записей вернуть
* `offset` - начиня с какой записи вернуть (сколько пропустить с начала) - нужно для огранизации постраничной навигации

Дополнительно:
* Данные для работы лежаит в файле `dataset.xml`
