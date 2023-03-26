import sys

DEFAULT_CLIENTS = 1

n_clients = DEFAULT_CLIENTS
if len(sys.argv) > 1 and sys.argv[1].isdigit():
    n_clients = int(sys.argv[1])

clients_string = ""

for i in range(1, n_clients+1):
    clients_string = clients_string + f'''
  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
'''
   
file_content = f'''version: '3.9'
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
  {clients_string}
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
'''

f = open("docker-compose-dev.yaml", "w")
f.write(file_content)
f.close()