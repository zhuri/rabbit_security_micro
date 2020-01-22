1. git clone https://github.com/zhuri/rabbit_micro.git

2. docker-compose build

3. You need two CLI tabs

4. In the first tab run -->> docker-compose exec -it publisher_c bash

5. Once you are inside the container run ./main binary executable

6. In the second tab run --->> docker-compose exec -it subscriber_c bash

7. Once you are inside the container run ./main binary executable

8. In the first tab your stdin will be transmitted to the second tab via rabbitmq

9. To see the state of the rabbitmq broker visit in your browser @ localhost:15672 

10. username: guest -- password:guest
