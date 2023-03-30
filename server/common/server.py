import socket
import logging
import signal
import json
from common import utils

BETS_REQUEST = 1
WINNERS_REQUEST = 2
BATCH_OK_RESPONSE = "OK"
WINNERS_OK_RESPONSE = 1
WINNERS_NOT_OK_RESPONSE = 0

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._agencies_stored = []
        self._lottery_draw_done = False
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
        try:
            data = client_sock.recv(1)
            request_type = int.from_bytes(data, "big")
            if request_type == BETS_REQUEST:
                self.__handle_bets(client_sock)
            elif request_type == WINNERS_REQUEST:
                self.__handle_winners(client_sock)
        except OSError as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
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

    def __handle_bets(self, client_sock):
        agency = 0
        while True:
            bets = self.__receive_batch(client_sock)
            if len(bets) == 0:
                break
            utils.store_bets(bets)
            agency = bets[0].agency
            self.__send_response(client_sock, BATCH_OK_RESPONSE)
        logging.info(f'action: store_bets | result: success | agency: {agency}')
        self._agencies_stored.append(agency)
        self.__lottery_draw()

    def __lottery_draw(self):
        if self._lottery_draw_done:
            return
        for i in range(1, 6):
            if i not in self._agencies_stored:
                return
        self._lottery_draw_done = True
        logging.info(f'action: lottery_draw | result: success')

    def __handle_winners(self, client_sock):
        if self._lottery_draw_done:
            response = WINNERS_OK_RESPONSE
        else:
            response = WINNERS_NOT_OK_RESPONSE

        response = response.to_bytes(1, "big")
        client_sock.send(response)

        if response == WINNERS_NOT_OK_RESPONSE:
            client_sock.close()
            return

        agency = client_sock.recv(2)
        agency = int.from_bytes(agency, "big")

        bets = utils.load_bets()
        winners = ""
        for bet in bets:
            if utils.has_won(bet) and bet.agency == agency:
                winners += bet.document
                winners += ","
        winners = winners[:-1]

        self.__send_response(client_sock, winners)

    def __receive_batch(self, client_sock) -> [utils.Bet]:
        total_length = client_sock.recv(2)
        total_length = int.from_bytes(total_length, "big")

        logging.debug(f'action: receive_length | result: success | msg: {total_length}')

        if total_length == 0:
            return []

        buffer = b''
        while len(buffer) < total_length:
            data = client_sock.recv(total_length - len(buffer))
            buffer += data

        json_batch = json.loads(buffer.decode("utf-8"))
        agency = json_batch["agency"]
        bets = []
        for json_bet in json_batch["bets"]:
            bet = utils.Bet(
                agency=agency,
                first_name=json_bet["first_name"],
                last_name=json_bet["last_name"],
                document=json_bet["document"],
                birthdate=json_bet["birthdate"],
                number=json_bet["number"],
            )
            bets.append(bet)

        addr = client_sock.getpeername()
        # logging.debug(f'action: receive_batch | result: success | ip: {addr[0]} | msg: {json_batch}')

        return bets

    def __send_response(self, client_sock, message):
        message += "\n"
        data = message.encode('utf-8')
        total_sent = 0
        while total_sent < len(data):
            sent = client_sock.send(data)
            total_sent += sent
        logging.debug(f'action: send_response | result: success | msg: {message}')

