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
      - CLI_LOG_LEVEL=INFO
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - type: bind
        source: ./client/config.yaml
        target: /config.yaml
      - type: bind
        source: ./.data/dataset/agency-{i}.csv
        target: /agency-{i}.csv
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
      - LOGGING_LEVEL=INFO
    networks:
      - testing_net
    volumes:
      - type: bind
        source: ./server/config.ini
        target: /config.ini
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