import socket
import logging
import signal
import json
from common import utils

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while True:
            try:
                client_sock = self.__accept_new_connection() 
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if e.errno == 9:  # Socket closed
                    break
                raise

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bet = self.__receive_bet(client_sock)
            utils.store_bets([bet])
            logging.info(f'action: store_bet | result: success | dni: {bet.document} | number: {bet.number}')
            self.__send_response(client_sock, "OK")

        except OSError as e:
            logging.error("action: receive_bet | result: fail | error: {e}")
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.debug('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.debug(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def __handle_sigterm(self, _a, _b):
        self._server_socket.close()
        logging.debug('action: close_server_socket | result: success')

    def __receive_bet(self, client_sock) -> utils.Bet:
        total_length = client_sock.recv(2)
        total_length = int.from_bytes(total_length, "big")

        buffer = b''
        while len(buffer) < total_length:
            data = client_sock.recv(total_length - len(buffer))
            buffer += data

        json_bet = json.loads(buffer.decode("utf-8"))
        addr = client_sock.getpeername()

        logging.debug(f'action: receive_bet | result: success | ip: {addr[0]} | msg: {json_bet}')

        return utils.Bet(
            agency=json_bet["agency"],
            first_name=json_bet["first_name"],
            last_name=json_bet["last_name"],
            document=json_bet["document"],
            birthdate=json_bet["birthdate"],
            number=json_bet["number"],
        )

    def __send_response(self, client_sock, message):
        message += "\n"
        data = message.encode('utf-8')
        total_sent = 0
        while total_sent < len(data):
            sent = client_sock.send(data)
            total_sent += sent
        logging.debug(f'action: send_response | result: success | msg: "OK"')

