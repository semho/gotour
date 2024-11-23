## Вопросы
1. ### Виртуализация.
   https://habr.com/ru/articles/657677/  
   https://otus.ru/journal/virtualizaciya-tipy-principy-raboty-osobennosti/
2. ### Гипервизор 1-2-3 типа.
   https://habr.com/ru/companies/vps_house/articles/349788/   
   https://market.cnews.ru/articles/2023-07-07_3_tipa_gipervizorov_i_ih_sravnenie   
   https://servermall.ru/blog/kakoy-gipervizor-vybrat/    
* **Тип 1 (bare-metal)** - работает напрямую с оборудованием. Этот тип гипервизоров работает на аппаратном уровне, без операционной системы хоста. Примеры: VMware ESXi, Microsoft Hyper-V и KVM
* **Тип 2 (hosted)** - работает поверх операционной системы, запускается как приложения внутри операционной системы (например, VirtualBox, VMware Workstation)
* **Тип 3 (гибридный)** - комбинирует особенности обоих типов. Например, гибридный гипервизор может иметь ядро, работающее на уровне железа, а некоторые компоненты управления могут работать на уровне операционной системы. Примеры гибридных: VMware Fusion и Parallels Desktop для Mac
3. ### Docker.
   https://habr.com/ru/companies/ruvds/articles/438796/  
   https://medium.com/webbdev/docker-bbb3de0f02c3
4. ### Контейнеризация.
   https://yandex.cloud/ru/docs/glossary/containerization    
   https://habr.com/ru/companies/otus/articles/767884/    
   https://eternalhost.net/blog/razrabotka/docker-kubernetes     


## Практика
### Задание
1. Реализовать для мессенджера storage типа PostgreSQL
2. Реализовать асинхронное сохранение сообщений в чат через Kafka
3. Спроектировать схему данных в Kafka при помощи protobuf
4. Сделать отдельный воркер для записи сообщений в БД
5. Различные инстансы соединить через docker-compose

Практику делаем в 8.less