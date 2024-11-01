## Вопросы
1. ### Брокеры сообщений. Очереди RabbitMQ / NATS / Redis   
   https://timeweb.cloud/tutorials/microservices/populyarnye-brokery-soobshchenij
   https://timeweb.cloud/tutorials/redis/broker-soobshchenij-redis
   https://university.ylab.io/articles/tpost/h819s4rn91-redis-rabbitmq-i-kafka-chto-vibrat-dlya

2. ### Различные структуры данных Redis.
   https://habr.com/ru/articles/144054/   

* **Строки (Strings)**. Это наиболее базовый тип данных в Redis. Они могут содержать любые данные, включая бинарные, и часто используются для хранения значений, таких как текст, сериализованные объекты или даже большие файлы в виде байтов.
* **Списки (Lists)**. Позволяют хранить упорядоченные последовательности строк. Redis реализует списки в виде связных списков, что делает их идеальными для операций с добавлением или удалением элементов на концах.
* **Множества (Sets)**. Это коллекции уникальных строк. Они поддерживают быстрые операции добавления, удаления и проверки наличия элемента, что делает их идеальными для работы с уникальными данными.
* **Хэш-таблицы (Hashes)**. Это структуры данных, ключами и значениями которых являются строки. Это позволяет эффективно представлять объекты (например, пользователей с полями и значениями).
* **Сортированные множества (Sorted Sets)**. Похожи на обычные множества, но каждый элемент связан с числовым «весом» или рейтингом. Эти элементы упорядочиваются в зависимости от данного веса.

3. ### Различные виды exchange RabbitMQ.
   https://habr.com/ru/articles/489086/   
* **Direct exchange**. Сообщения доставляются в очереди на основе ключа маршрутизации сообщения.
* **Fanout exchange**. Сообщения доставляются во все привязанные к нему очереди, даже если в сообщении задан ключ маршрутизации.
* **Topic exchange**. Для публикации сообщений в очередь происходит подстановочное совпадение ключа маршрутизации и шаблона маршрутизации, указанного в привязке.
* **Headers exchange**. Для маршрутизации используются атрибуты заголовка сообщения. 
4. ### Kafka. Топики Kafka. Партиции. Consumer / Consumer Group
   https://habr.com/ru/articles/466585/    
   https://slurm.io/blog/tpost/pnyjznpvr1-apache-kafka-osnovi-tehnologii
5. ### Различные семантики доставки. Transactional Outbox / Transactional Inbox.
   https://softwaremill.com/microservices-101/
5. ### Паттерны проектирования Event Sourcing. Messaging.
   https://bool.dev/blog/detail/pattern-cqrs-i-event-sourcing  
   https://en.wikipedia.org/wiki/Messaging_pattern
   https://dev.to/simsekahmett/messaging-patterns-101-a-comprehensive-guide-for-software-developers-2j3c
   
 